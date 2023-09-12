package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"hacompanion/api"
	"hacompanion/entity"
	"hacompanion/sensor"
	"hacompanion/util"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
)

var (
	AppID        = "homeassistant-desktop-companion"
	AppName      = "Home Assistant Desktop Companion"
	Manufacturer = "https://github.com/tobias-kuendig/hacompanion"
	Model        = "hacompanion"
	OsName       = runtime.GOOS
	Version      = "1.0.8"
	// NOTE for Home Assistant 2022.3.3 and earlier versions:
	// Because OsVersion populates the "Firmware" field on the devices page
	// and the Companion version is not displayed there otherwise,
	// we construct OsVersion with both Version and OsName.
	OsVersion = fmt.Sprintf("%s (%s)", Version, OsName)
)

// Kernel holds all of the application's dependencies.
type Kernel struct {
	config        *Config
	api           *api.API
	notifications *NotificationServer
	ctxCancel     context.CancelFunc
	bgProcesses   *sync.WaitGroup
}

func main() {
	var hassToken string
	var hassHost string
	var deviceName string
	var configFlag string
	flag.StringVar(&configFlag, "config", "~/.config/hacompanion.toml", "Path to the config file")
	flag.StringVar(&hassHost, "host", "", "Home Assistant host")
	flag.StringVar(&hassToken, "token", "", "Long-lived access token")
	flag.StringVar(&deviceName, "device-name", "", "Device name")
	flag.Parse()

	configFile, err := homePathFromString(configFlag)
	if err != nil {
		log.Fatalf("failed to parse config flag %s: %s", configFlag, err)
	}
	if exists, _ := util.FileExists(configFile.Path); !exists {
		log.Fatalf("could not load config file %s", configFile.Path)
	}

	// Get some randomness going.
	rand.Seed(time.Now().UnixNano())

	// Try to parse the config file.
	var config Config
	b, err := os.ReadFile(configFile.Path)
	if err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}
	if _, err = toml.Decode(string(b), &config); err != nil {
		log.Fatalf("failed to parse config file: %s", err)
	}

	// # Home Assistant Host
	//
	// hassHost is set by searching the following in order,
	// using the first found value.
	//
	//  1. The "-host" command line flag.
	//  2. The "HASS_HOST" environment variable.
	//  3. The "homeassistant.host" config file value.
	//  4. The default value "http://homeassistant.local:8123".
	if hassHost == "" {
		hassHost = os.Getenv("HASS_HOST")
		if hassHost == "" {
			hassHost = config.HomeAssistant.Host
		}
		if hassHost == "" {
			hassHost = "http://homeassistant.local:8123"
		}
	}

	// # Home Assistant Token
	//
	// hassToken is set by searching the following in order,
	// using the first found value.
	//
	//  1. The "-token" command line flag.
	//  2. The "HASS_TOKEN" environment variable.
	//  3. The "homeassistant.token" config file value.
	if hassToken == "" {
		hassToken = os.Getenv("HASS_TOKEN")
		if hassToken == "" {
			hassToken = config.HomeAssistant.Token
		}
	}

	// # Device Name
	//
	// deviceName is set by searching the following in order,
	// using the first found value.
	//
	//  1. The "-device-name" command line flag.
	//  2. The "HASS_DEVICE_NAME" environment variable.
	//  3. The "homeassistant.device_name" config file value.
	//  4. The system hostname.
	if deviceName == "" {
		deviceName = os.Getenv("HASS_DEVICE_NAME")
		if deviceName == "" {
			deviceName = config.HomeAssistant.DeviceName
		}
		if deviceName == "" {
			hostname, err := os.Hostname()
			if err != nil {
				log.Fatalf("failed to determine hostname: %s. Please set device name via HASS_DEVICE_NAME or the config file", err)
			}

			deviceName = hostname
		}
	}

	// Build the application kernel.
	k := Kernel{
		config: &config,
		api:    api.NewAPI(hassHost, hassToken, deviceName),
	}

	// Start the main process.
	go func() {
		err = k.Run(context.Background())
		if err != nil {
			log.Fatalf("failed to start application: %s", err)
		}
	}()

	// Handle shutdowns gracefully.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Wait for the shutdown signal.
	<-stop

	// Give the application a few seconds to shut down.
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	if err = k.Shutdown(ctx); err != nil {
		log.Fatalf("application shutdown error: %s\n", err)
	} else {
		log.Println("application stopped")
	}
}

// Run runs the application.
func (k *Kernel) Run(appCtx context.Context) error {
	log.Printf("Starting Desktop Companion version %s", Version)
	// Create a global application context that is later used for proper shutdowns.
	ctx, cancel := context.WithCancel(appCtx)
	k.ctxCancel = cancel

	// Create a WaitGroup for all backend processes.
	k.bgProcesses = &sync.WaitGroup{}

	// Make sure the device is registered in Home Assistant.
	registration, err := k.getRegistration(ctx)
	if err != nil {
		return fmt.Errorf("failed to get device registration: %w", err)
	}
	k.api.Registration = registration

	// Update device registration data.
	err = k.updateRegistration(ctx, registration)
	if err != nil {
		// Log error and continue, this shouldn't be fatal.
		fmt.Println("failed to update device registration info: %w", err)
	}

	// Parse out all sensors from the config file and register them in Home Assistant.
	sensors, err := k.buildSensors(k.config)
	if err != nil {
		return fmt.Errorf("failed to build sensors from config: %w", err)
	}
	err = k.api.RegisterSensors(ctx, sensors)
	if err != nil {
		return err
	}

	// Start the notifications server.
	k.notifications = NewNotificationServer(registration, k.config.Notifications.Listen)
	go k.notifications.Listen(ctx)

	// The Companion instance gathers sensor data and forwards it to Home Assistant.
	c := NewCompanion(k.api, sensors)

	// Start the background processes.
	k.bgProcesses.Add(1)
	go c.RunBackgroundProcesses(ctx, k.bgProcesses)

	// Run the first update immediately on startup.
	c.UpdateSensorData(ctx)

	// Keep updating the sensor data in a regular interval,
	// until the application context gets cancelled.
	t := time.NewTicker(k.config.Companion.UpdateInterval.Duration)

	for {
		select {
		case <-t.C:
			c.UpdateSensorData(ctx)
		case <-ctx.Done():
			return nil
		}
	}
}

// Shutdown ends the main routine gracefully.
func (k *Kernel) Shutdown(ctx context.Context) error {
	// Cancel global context, then wait for all background processes to quit.
	k.ctxCancel()

	done := make(chan struct{})
	go func() {
		k.bgProcesses.Wait()
		close(done)
	}()

	// Stop the notification server.
	if err := k.notifications.Server.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	// Wait for either everything to shut down properly
	// or the context timeout to be reached.
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		break
	}

	return nil
}

// buildSensors returns a slice of concrete Sensor types based on the configuration.
func (k *Kernel) buildSensors(config *Config) ([]entity.Sensor, error) {
	var sensors []entity.Sensor
	// Parse default sensor configuration.
	for key, sensorConfig := range config.Sensors {
		if !sensorConfig.Enabled {
			continue
		}
		definition, ok := sensorDefinitions[key]
		if !ok {
			return nil, fmt.Errorf("unknown sensor %s in config", key)
		}
		data := definition(sensorConfig.Meta)
		sensors = append(sensors, entity.Sensor{
			Type:        data.Type,
			Name:        sensorConfig.Name,
			UniqueID:    key,
			Runner:      data.Runner(sensorConfig.Meta),
			DeviceClass: data.DeviceClass,
			Icon:        data.Icon,
			Unit:        data.Unit,
		})
	}
	// Parse custom scripts.
	for key, scriptConfig := range config.Script {
		sensors = append(sensors, entity.Sensor{
			Type:        scriptConfig.Type,
			Runner:      sensor.NewScriptRunner(scriptConfig),
			DeviceClass: scriptConfig.DeviceClass,
			Icon:        scriptConfig.Icon,
			Name:        scriptConfig.Name,
			UniqueID:    key,
			Unit:        scriptConfig.UnitOfMeasurement,
		})
	}
	return sensors, nil
}

// getRegistration tries to read an existing Home Assistant device registration.
// If it does not exist, it registers a new device with Home Assistant.
func (k *Kernel) getRegistration(ctx context.Context) (api.Registration, error) {
	var registration api.Registration
	var err error
	_, err = os.Stat(k.config.Companion.RegistrationFile.Path)
	// If there is a registration file available, use it.
	if err == nil {
		var b []byte
		b, err = os.ReadFile(k.config.Companion.RegistrationFile.Path)
		if err != nil {
			return registration, err
		}
		// TODO: Try to register device in the case of invalid json (e.g. empty/corrupted file),
		// rather than simply returning the error.
		err = json.Unmarshal(b, &registration)
		return registration, err
	}
	// Something went wrong, return the error.
	if !errors.Is(err, fs.ErrNotExist) {
		return registration, err
	}
	// No registration file found, try to register device.
	return k.registerDevice(ctx)
}

// registerDevice registers a new device with Home Assistant.
func (k *Kernel) registerDevice(ctx context.Context) (api.Registration, error) {
	id := util.RandomString(8)
	token := util.RandomString(8)

	pushURL, err := k.config.GetPushURL()
	if err != nil {
		log.Println("Push notifications will not work with your current config")
	}

	registration, err := k.api.RegisterDevice(ctx, api.RegisterDeviceRequest{
		AppData: api.AppData{
			PushToken: token,
			PushURL:   pushURL,
		},
		AppID:              AppID,
		AppName:            AppName,
		AppVersion:         Version,
		DeviceID:           id,
		DeviceName:         k.api.DeviceName,
		Manufacturer:       Manufacturer,
		Model:              Model,
		OsVersion:          OsVersion,
		SupportsEncryption: false,
	})
	if err != nil {
		return registration, err
	}
	registration.PushToken = token
	// Parse the response and save it to the filesystem.
	j, err := registration.JSON()
	if err != nil {
		return registration, err
	}
	err = os.MkdirAll(filepath.Dir(k.config.Companion.RegistrationFile.Path), 0700)
	if err != nil {
		return registration, err
	}
	err = os.WriteFile(k.config.Companion.RegistrationFile.Path, j, 0600)
	if err != nil {
		return registration, err
	}
	return registration, err
}

// updateRegistration updates app registration data.
func (k *Kernel) updateRegistration(ctx context.Context, registration api.Registration) error {
	pushURL, err := k.config.GetPushURL()
	if err != nil {
		log.Println("Push notifications will not work with your current config")
	}
	err = k.api.UpdateRegistration(ctx, api.UpdateRegistrationRequest{
		AppData: api.AppData{
			PushToken: registration.PushToken,
			PushURL:   pushURL,
		},
		AppVersion:   Version,
		DeviceName:   k.api.DeviceName,
		Manufacturer: Manufacturer,
		Model:        Model,
		OsVersion:    OsVersion,
	})
	return err
}

// NullRunner is a Runner that does not do anything.
type NullRunner struct{}

func (n NullRunner) Run(ctx context.Context) (*entity.Payload, error) { return nil, nil }

// duration is used to unmarshal text durations into a time.Duration.
type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// homePath enables support for ~/home/paths.
type homePath struct {
	Path string
}

func homePathFromString(in string) (*homePath, error) {
	h := &homePath{}
	err := h.UnmarshalText([]byte(in))
	return h, err
}

func (h *homePath) UnmarshalText(text []byte) error {
	h.Path = string(text)
	if strings.HasPrefix(h.Path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		h.Path = filepath.Join(usr.HomeDir, string(text[2:]))
		return nil
	}
	return nil
}
