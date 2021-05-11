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
		if output.payload == nil {
			continue
		}
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
		log.Printf("failed to update sensor data: %s", err)
	}
}

// RunBackgroundProcesses starts all background processes.
func (c *Companion) RunBackgroundProcesses(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	var processWg sync.WaitGroup

	go c.UpdateCompanionRunningState(ctx, &processWg)

	processWg.Wait()
}

// UpdateCompanionRunningState updates the companion running state.
func (c *Companion) UpdateCompanionRunningState(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	update := func(state bool) {
		err := c.api.UpdateSensorData(ctx, []UpdateSensorDataRequest{{
			State:    state,
			Type:     "sensor",
			Icon:     "mdi:heart-pulse",
			UniqueId: "companion_running",
		}})
		if err != nil {
			log.Printf("failed to update companion_running state: %s", err)
		}
	}

	update(true)
	defer func() {
		update(false)
		wg.Done()
	}()

	<-ctx.Done()
}
