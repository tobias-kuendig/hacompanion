package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
)

var registrationFile = "registration.json"

func main() {
	// Try to parse the config file.
	var config Config
	b, err := ioutil.ReadFile("hadaemon.toml")
	if err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}
	if _, err = toml.Decode(string(b), &config); err != nil {
		log.Fatalf("failed to parse config file: %s", err)
	}
	// Get some randomness going.
	rand.Seed(time.Now().UnixNano())
	// Handle shutdowns gracefully.
	sigs := make(chan os.Signal, 1)
	done := make(chan error, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context once a shutdown signal was received.
	go func() {
		<- sigs
		log.Println("SIG RECEIVED")
		cancel()
	}()

	// Start the main process.
	go func(ctx context.Context) {
		api := NewAPI(config.Host, config.Token)
		err := run(ctx, config, api)
		log.Printf("ERR RECEIVED: %s\n", err)
		done <- err
	}(ctx)

	// Wait for the error to return.
	if err := <-done; err != nil {
		log.Fatalf("exited with error: %s", err)
	}
	log.Println("DONE")
}

func run(ctx context.Context, config Config, api *API) error {
	// Make sure the device is registered in Home Assistant.
	registration, err := getRegistration(ctx, api, config)
	if err != nil {
		return fmt.Errorf("failed to get device registrationo: %w", err)
	}
	api.registration = registration

	// Parse all sensors out of the config file and register them in Home Assistant.
	sensors, err := buildSensors(config)
	if err != nil {
		return fmt.Errorf("failed to build sensors from config: %w", err)
	}
	err = api.RegisterSensors(ctx, sensors)
	if err != nil {
		return err
	}

	// Start the notifications server.
	notifications := NewNotificationServer(registration)
	go notifications.Listen(ctx)

	// The Companion gathers sensor data and forwards it to Home Assistant.
	companion := NewCompanion(api, sensors)

	// Start the background processes.
	var bgProcesses sync.WaitGroup
	go companion.RunBackgroundProcesses(ctx, &bgProcesses)

	// Run the first update immediately.
	companion.UpdateSensorData(ctx)

	// Keep updating the sensor data in a regular interval.
	t := time.NewTicker(10 * time.Second)

	// Done chan signals, when everything is done.
	done := make(chan struct{})
	go func() {
		bgProcesses.Wait()
		done <- struct{}{}
	}()

	for {
		select {
		case <-t.C:
			companion.UpdateSensorData(ctx)
		case <-ctx.Done():
			select {
			case <-done:
				return nil
			case <-time.After(5 * time.Second):
				return errors.New("failed to shut down background processes")
			}
		}
	}
}

type sensorDefinition struct {
	runner      func(Meta) runner
	deviceClass string
	icon        string
	unit        string
}

var sensorDefinitions = map[string]func(m Meta) sensorDefinition{
	"cpu_temp": func(m Meta) sensorDefinition {
		unit := "C"
		if !m.GetBool("celsius") {
			unit = "F"
		}
		return sensorDefinition{
			runner:      func(m Meta) runner { return NewCPUTemp(m) },
			deviceClass: "temperature",
			icon:        "mdi:thermometer",
			unit:        unit,
		}
	},
	"cpu_usage": func(m Meta) sensorDefinition {
		return sensorDefinition{
			runner: func(m Meta) runner { return NewCPUUsage() },
			icon:   "mdi:gauge",
			unit:   "%",
		}
	},
	"memory": func(m Meta) sensorDefinition {
		return sensorDefinition{
			runner: func(Meta) runner { return NewMemory() },
			icon:   "mdi:memory",
		}
	},
	"power": func(m Meta) sensorDefinition {
		return sensorDefinition{
			runner:      func(m Meta) runner { return NewPower(m) },
			icon:        "mdi:battery",
			deviceClass: "battery",
			unit:        "%",
		}
	},
	"uptime": func(m Meta) sensorDefinition {
		return sensorDefinition{
			runner:      func(Meta) runner { return NewUptime() },
			icon:        "mdi:sort-clock-descending",
			deviceClass: "timestamp",
			unit:        "ISO8601",
		}
	},
	"load_avg": func(m Meta) sensorDefinition {
		return sensorDefinition{
			runner: func(Meta) runner { return NewLoadAVG() },
			icon:   "mdi:gauge",
		}
	},
	"webcam": func(m Meta) sensorDefinition {
		return sensorDefinition{
			runner: func(Meta) runner { return NewWebCam() },
			icon:   "mdi:webcam",
		}
	},
	"audio_volume": func(m Meta) sensorDefinition {
		return sensorDefinition{
			runner: func(Meta) runner { return NewAudioVolume() },
			icon:   "mdi:volume-high",
			unit:   "%",
		}
	},
	"online_check": func(m Meta) sensorDefinition {
		return sensorDefinition{
			runner: func(m Meta) runner { return NewOnlineCheck(m) },
			icon:   "mdi:shield-check-outline",
		}
	},
	"companion_running": func(m Meta) sensorDefinition {
		return sensorDefinition{
			runner: func(Meta) runner { return NullRunner{} },
			icon:   "mdi:heart-pulse",
		}
	},
}

type Sensor struct {
	runner      runner
	deviceClass string
	icon        string
	name        string
	uniqueID    string
	unit        string
}

func (s Sensor) String() string {
	return fmt.Sprintf("%s (%s)", s.name, s.uniqueID)
}

func buildSensors(config Config) ([]Sensor, error) {
	var sensors []Sensor
	for key, sensorConfig := range config.Sensors {
		if !sensorConfig.Enabled {
			continue
		}
		definition, ok := sensorDefinitions[key]
		if !ok {
			return nil, fmt.Errorf("unknown sensor %s in config", key)
		}
		data := definition(sensorConfig.Meta)
		sensors = append(sensors, Sensor{
			name:        sensorConfig.Name,
			uniqueID:    key,
			runner:      data.runner(sensorConfig.Meta),
			deviceClass: data.deviceClass,
			icon:        data.icon,
			unit:        data.unit,
		})
	}
	return sensors, nil
}

func getRegistration(ctx context.Context, api *API, config Config) (Registration, error) {
	var registration Registration
	var err error
	_, err = os.Stat(registrationFile)
	// If there is a registration file available, use it.
	if err == nil {
		var b []byte
		b, err = ioutil.ReadFile(registrationFile)
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
	return registerDevice(ctx, api, config)
}

func registerDevice(ctx context.Context, api *API, config Config) (Registration, error) {
	id := RandomString(8)
	token := RandomString(8)
	registration, err := api.RegisterDevice(ctx, RegisterDeviceRequest{
		DeviceID:           id,
		AppID:              "hadaemon",
		AppName:            "Home Assistant Daemon",
		AppVersion:         "1.0",
		DeviceName:         config.DeviceName,
		SupportsEncryption: false,
		AppData: AppData{
			PushToken: token,
			PushURL:   "http://192.168.1.2:8080/notification",
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
	err = ioutil.WriteFile(registrationFile, j, 0600)
	if err != nil {
		return registration, err
	}
	return registration, err
}

type Config struct {
	DeviceName string                  `toml:"device_name"`
	Prefix     string                  `toml:"prefix"`
	Token      string                  `toml:"token"`
	Host       string                  `toml:"host"`
	Sensors    map[string]SensorConfig `toml:"sensor"`
}

type SensorConfig struct {
	Enabled bool
	Name    string
	Meta    map[string]interface{}
}

type runner interface {
	run(ctx context.Context) (*payload, error)
}

type Output struct {
	payload *payload
	sensor  Sensor
}

type Meta map[string]interface{}

func (m Meta) GetBool(key string) bool {
	if v, ok := m[key]; ok {
		if v == true {
			return true
		}
		return false
	}
	return false
}

func (m Meta) GetString(key string) string {
	if v, ok := m[key]; ok {
		if value, isString := v.(string); isString {
			return value
		}
	}
	return ""
}

func (s Sensor) update(ctx context.Context, wg *sync.WaitGroup, outputs *Outputs) {
	defer wg.Done()
	value, err := s.runner.run(ctx)
	if err != nil {
		log.Printf("failed to run sensor %s: %s", s, err)
		return
	}
	log.Printf("received payload for %s: %+v", s.uniqueID, value)
	outputs.Add(Output{sensor: s, payload: value})
}

func respondError(w http.ResponseWriter, error string, status int) {
	var resp struct {
		Error string `json:"errorMessage"`
	}
	resp.Error = error
	b, err := json.Marshal(resp)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, string(b))))
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(status)
	w.Write(b)
}

// Home Assistant errors without a rateLimits response.
// There is no rate limiting implemented on our side so we return a dummy response.
func respondSuccess(w http.ResponseWriter) {
	w.Write([]byte(`
		{
			"rateLimits": {
				"successful": 1,
				"errors": 0,
				"maximum": 150,
				"resetsAt": "2019-04-08T00:00:00.000Z"
			}
		}
	`))
}

type NullRunner struct{}

func (n NullRunner) run(ctx context.Context) (*payload, error) { return nil, nil }
