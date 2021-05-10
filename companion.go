package main

import (
	"context"
	"log"
	"sync"
)

type Outputs struct {
	data []*Output
	lock sync.Mutex
}

func NewOutputs() Outputs {
	return Outputs{
		lock: sync.Mutex{},
		data: make([]*Output, 0),
	}
}

func (o *Outputs) Add(output Output) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.data = append(o.data, &output)
}

type Companion struct {
	sensors []Sensor
	api     *API
	wg      sync.WaitGroup
}

func NewCompanion(api *API, sensors []Sensor) *Companion {
	return &Companion{
		api:     api,
		sensors: sensors,
	}
}

func (c *Companion) UpdateSensorData(ctx context.Context) {
	outputs := NewOutputs()

	// Fetch all sensor values in parallel.
	for _, sensor := range c.sensors {
		c.wg.Add(1)
		go sensor.update(ctx, &c.wg, &outputs)
	}

	c.wg.Wait()

	// Build one request to send all updated values to Home Assistant.
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

	err := c.api.UpdateSensorData(ctx, data)
	if err != nil {
		log.Printf("failed to update sensor data: %w", err)
	}
}
