package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"hacompanion/api"
	"hacompanion/companion"
	"hacompanion/entity"
	"hacompanion/util"
)

// Version contains the binary's release version.
var Version = "1.0"

// Kernel holds all of the application's dependencies.
type Kernel struct {
	config        *Config
	api           *api.API
	notifications *companion.NotificationServer
	ctxCancel     context.CancelFunc
	bgProcesses   *sync.WaitGroup
}

func main() {
	var configFlag string
	flag.StringVar(&configFlag, "config", "~/.config/hacompanion.toml", "Path to the config file")
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
	b, err := ioutil.ReadFile(configFile.Path)
	if err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}
	if _, err = toml.Decode(string(b), &config); err != nil {
		log.Fatalf("failed to parse config file: %s", err)
	}

	// Build the application kernel.
	k := Kernel{
		config: &config,
		api:    api.NewAPI(config.HomeAssistant.Host, config.HomeAssistant.Token),
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
	log.Printf("Starting companion version %s", Version)
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

	// Parse all sensors out of the config file and register them in Home Assistant.
	sensors, err := k.buildSensors(k.config)
	if err != nil {
		return fmt.Errorf("failed to build sensors from config: %w", err)
	}
	err = k.api.RegisterSensors(ctx, sensors)
	if err != nil {
		return err
	}

	// Start the notifications server.
	k.notifications = companion.NewNotificationServer(registration, k.config.Notifications.Listen)
	go k.notifications.Listen(ctx)

	// The Companion gathers sensor data and forwards it to Home Assistant.
	c := companion.NewCompanion(k.api, sensors)

	// Start the background processes.
	k.bgProcesses.Add(1)
	go c.RunBackgroundProcesses(ctx, k.bgProcesses)

	// Run the first update immediately.
	c.UpdateSensorData(ctx)

	// Keep updating the sensor data in a regular interval until
	// the application context gets cancelled.
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

// Shutdown shuts down the main routine gracefully.
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

	// Wait for either everything to shut down properly or the the timeout of the context to exceed.
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		break
	}

	return nil
}

// buildSensor returns a slice of concrete Sensor types based on the configuration.
func (k *Kernel) buildSensors(config *Config) ([]entity.Sensor, error) {
	var sensors []entity.Sensor
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
			Name:        sensorConfig.Name,
			UniqueID:    key,
			Runner:      data.Runner(sensorConfig.Meta),
			DeviceClass: data.DeviceClass,
			Icon:        data.Icon,
			Unit:        data.Unit,
		})
	}
	return sensors, nil
}

// getRegistration tries to read an existing Home Assistant device registration.
// If it does not exist, it register a new device with Home Assistant.
func (k *Kernel) getRegistration(ctx context.Context) (api.Registration, error) {
	var registration api.Registration
	var err error
	_, err = os.Stat(k.config.Companion.RegistrationFile.Path)
	// If there is a registration file available, use it.
	if err == nil {
		var b []byte
		b, err = ioutil.ReadFile(k.config.Companion.RegistrationFile.Path)
		if err != nil {
			return registration, err
		}
		err = json.Unmarshal(b, &registration)
		return registration, err
	}
	// Something went wrong, return the error.
	if !os.IsNotExist(err) {
		return registration, err
	}
	return k.registerDevice(ctx)
}

// registerDevices registers a new device with Home Assistant.
func (k *Kernel) registerDevice(ctx context.Context) (api.Registration, error) {
	id := util.RandomString(8)
	token := util.RandomString(8)
	registration, err := k.api.RegisterDevice(ctx, api.RegisterDeviceRequest{
		DeviceID:           id,
		AppID:              "homeassistant-desktop-companion",
		AppName:            "Home Assistant Desktop Companion",
		AppVersion:         Version,
		DeviceName:         k.config.HomeAssistant.DeviceName,
		SupportsEncryption: false,
		AppData: api.AppData{
			PushToken: token,
			PushURL:   k.config.Notifications.PushURL,
		},
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
	err = ioutil.WriteFile(k.config.Companion.RegistrationFile.Path, j, 0600)
	if err != nil {
		return registration, err
	}
	return registration, err
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
