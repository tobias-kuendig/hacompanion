package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

var registrationFile = "registration.json"

func main() {
	var config Config
	b, err := ioutil.ReadFile("hadaemon.toml")
	if err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}
	if _, err = toml.Decode(string(b), &config); err != nil {
		log.Fatalf("failed to parse config file: %s", err)
	}

	api := NewAPI(config.Host, config.Token)
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())

	registration, err := getRegistration(ctx, api, config)
	if err != nil {
		log.Fatalf("failed to get device registration: %s", err)
	}
	api.registration = registration

	sensors, err := buildSensors(config)
	if err != nil {
		log.Fatalf("failed to build sensors from config: %s", err)
	}

	for _, sensor := range sensors {
		err = api.RegisterSensor(ctx, RegisterSensorRequest{
			Type:              "sensor",
			DeviceClass:       sensor.deviceClass,
			Icon:              sensor.icon,
			Name:              sensor.name,
			UniqueId:          sensor.uniqueID,
			UnitOfMeasurement: sensor.unit,
		})
		if err != nil {
			log.Fatalf("failed to register sensor: %s", err)
		}
	}

	var wg sync.WaitGroup
	outputs := Outputs{
		lock: sync.Mutex{},
		data: make([]*Output, 0),
	}
	for _, sensor := range sensors {
		wg.Add(1)
		go sensor.start(ctx, &wg, &outputs)
	}
	wg.Wait()

	var data []UpdateSensorDataRequest
	for _, output := range outputs.data {
		data = append(data, UpdateSensorDataRequest{
			Type:       "sensor",
			State:      output.payload.State,
			Attributes: output.payload.Attributes,
			UniqueId:   output.sensor.uniqueID,
			Icon:       output.sensor.icon,
		})
	}
	err = api.UpdateSensorData(ctx, data)
	if err != nil {
		log.Fatalf("failed to update sensor data: %s", err)
	}
	os.Exit(2)
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
		if !m.GetBool("celcius") {
			unit = "F"
		}
		return sensorDefinition{
			runner:      func(m Meta) runner { return NewCPUTemp(m) },
			deviceClass: "temperature",
			icon:        "mdi:thermometer",
			unit:        unit,
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
}

type Sensor struct {
	runner      runner
	deviceClass string
	icon        string
	name        string
	uniqueID    string
	unit        string
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
	registration, err := api.RegisterDevice(ctx, RegisterDeviceRequest{
		DeviceID:           id,
		AppID:              "hadaemon",
		AppName:            "Home Assistant Daemon",
		AppVersion:         "1.0",
		DeviceName:         config.DeviceName,
		SupportsEncryption: false,
	})
	if err != nil {
		return registration, err
	}
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

type Outputs struct {
	data []*Output
	lock sync.Mutex
}

func (o *Outputs) Add(output Output) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.data = append(o.data, &output)
}

func (i Sensor) start(ctx context.Context, wg *sync.WaitGroup, outputs *Outputs) {
	defer wg.Done()
	value, err := i.runner.run(ctx)
	if err != nil {
		log.Printf("failed to run sensor %s: %s", i, err)
		return
	}
	log.Printf("received payload: %+v", value)
	outputs.Add(Output{sensor: i, payload: value})
}
