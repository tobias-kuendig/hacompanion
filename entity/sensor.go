package entity

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// Runner is used to gather data of any kind.
type Runner interface {
	Run(ctx context.Context) (*Payload, error)
}

// SensorDefinition contains all Home Assistant attributes.
type SensorDefinition struct {
	Type        string
	Runner      func(Meta) Runner
	DeviceClass string
	Icon        string
	Unit        string
}

// SensorConfig contains the configuration for a single sensor.
type SensorConfig struct {
	Enabled bool
	Name    string
	Meta    map[string]interface{}
}

// ScriptConfig contains the definition of a custom script sensor.
type ScriptConfig struct {
	Path              string
	Name              string
	Icon              string
	Type              string
	UnitOfMeasurement string `toml:"unit_of_measurement"`
	DeviceClass       string `toml:"device_class"`
}

// Sensor is a concrete instance of a sensor defined in the config file.
// Its Runner is run to gather data.
type Sensor struct {
	Type        string
	Runner      Runner
	DeviceClass string
	Icon        string
	Name        string
	UniqueID    string
	Unit        string
}

func (s Sensor) String() string {
	return fmt.Sprintf("%s (%s)", s.Name, s.UniqueID)
}

// Update runs a Sensor's Runner and returns the outputs.
func (s Sensor) Update(ctx context.Context, wg *sync.WaitGroup, outputs *Outputs) {
	defer wg.Done()
	value, err := s.Runner.Run(ctx)
	if err != nil {
		log.Printf("failed to run sensor %s: %s", s, err)
		return
	}
	log.Printf("received Payload for %s: %+v", s.UniqueID, value)
	outputs.Add(Output{Sensor: s, Payload: value})
}

// Invalidate sets the state of the given sensor(s) to unavailable.
func (s Sensor) Invalidate(outputs *Outputs) {
	p := NewPayload()
	p.State = "unavailable"
	log.Printf("received p for %s: %+v", s.UniqueID, p)
	outputs.Add(Output{Sensor: s, Payload: p})
}
