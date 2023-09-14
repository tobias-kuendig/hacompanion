package main

import (
	"context"
	"hacompanion/api"
	"hacompanion/entity"
	"log"
	"sync"
)

type Companion struct {
	sensors []entity.Sensor
	api     *api.API
	wg      sync.WaitGroup
}

func NewCompanion(api *api.API, sensors []entity.Sensor) *Companion {
	return &Companion{
		api:     api,
		sensors: sensors,
	}
}

func (c *Companion) UpdateSensorData(ctx context.Context) {
	outputs := entity.NewOutputs()

	// Fetch all sensor values in parallel.
	for _, sensor := range c.sensors {
		c.wg.Add(1)
		go sensor.Update(ctx, &c.wg, &outputs)
	}

	c.wg.Wait()

	// Build one request to send all updated values to Home Assistant.
	var data []api.UpdateSensorDataRequest
	for _, output := range outputs.Data {
		if output.Payload == nil {
			continue
		}
		icon := output.Payload.Icon
		if icon == "" {
			icon = output.Sensor.Icon
		}
		data = append(data, api.UpdateSensorDataRequest{
			Type:       output.Sensor.Type,
			State:      output.Payload.State,
			Attributes: output.Payload.Attributes,
			UniqueID:   output.Sensor.UniqueID,
			Icon:       icon,
		})
	}

	err := c.api.UpdateSensorData(ctx, data)
	if err != nil {
		log.Printf("failed to update sensor data: %s", err)
	}
}

// RunBackgroundProcesses starts all background processes.
func (c *Companion) RunBackgroundProcesses(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	var processWg sync.WaitGroup
	processWg.Add(1)

	go c.UpdateCompanionRunningState(ctx, &processWg)

	processWg.Wait()
}

// UpdateCompanionRunningState updates the Companion running state.
func (c *Companion) UpdateCompanionRunningState(ctx context.Context, wg *sync.WaitGroup) {
	update := func(state bool) {
		bgCtx := context.Background()
		if !state {
			log.Printf("Invalidating all sensors")
			c.InvalidateAllSensors(bgCtx)
		}
		err := c.api.UpdateSensorData(bgCtx, []api.UpdateSensorDataRequest{{
			State:    state,
			Type:     "binary_sensor",
			Icon:     "mdi:heart-pulse",
			UniqueID: "companion_running",
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

func (c *Companion) InvalidateAllSensors(ctx context.Context) {
	outputs := entity.NewOutputs()

	// Invalidate every registered sensor.
	for _, sensor := range c.sensors {
		sensor.Invalidate(&outputs)
	}

	// Build one request to send all updated values to Home Assistant.
	var data []api.UpdateSensorDataRequest
	for _, output := range outputs.Data {
		if output.Payload == nil {
			continue
		}
		data = append(data, api.UpdateSensorDataRequest{
			Type:     output.Sensor.Type,
			State:    output.Payload.State,
			UniqueID: output.Sensor.UniqueID,
			Icon:     output.Sensor.Icon,
		})
	}

	err := c.api.UpdateSensorData(ctx, data)
	if err != nil {
		log.Printf("failed to update sensor data: %s", err)
	}
}
