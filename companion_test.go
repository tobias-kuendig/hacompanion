package main

import (
	"testing"

	"hacompanion/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildUpdateSensorDataRequestsIncludesStateClass(t *testing.T) {
	outputs := entity.NewOutputs()
	outputs.Add(entity.Output{
		Sensor: entity.Sensor{
			Type:       "sensor",
			UniqueID:   "cpu_usage",
			Icon:       "mdi:gauge",
			StateClass: "measurement",
		},
		Payload: &entity.Payload{
			State:      42,
			Icon:       "mdi:cpu-64-bit",
			Attributes: map[string]interface{}{"foo": "bar"},
		},
	})

	data := buildUpdateSensorDataRequests(outputs, true)

	require.Len(t, data, 1)
	assert.Equal(t, "measurement", data[0].StateClass)
	assert.Equal(t, "cpu_usage", data[0].UniqueID)
	assert.Equal(t, "mdi:cpu-64-bit", data[0].Icon)
	assert.Equal(t, 42, data[0].State)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, data[0].Attributes)
}
