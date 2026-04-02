package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"hacompanion/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestRegisterSensorsIncludesStateClass(t *testing.T) {
	var payload registerSensorRequestPayload
	client := NewAPI("http://example.com", "token", "device", true)
	client.Registration = Registration{WebhookID: "abc123"}
	client.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &payload))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{}")),
			Header:     make(http.Header),
		}, nil
	})

	err := client.RegisterSensors(context.Background(), []entity.Sensor{{
		Type:       "sensor",
		Name:       "CPU Usage",
		UniqueID:   "cpu_usage",
		Icon:       "mdi:gauge",
		StateClass: "measurement",
		Unit:       "%",
	}})
	require.NoError(t, err)

	assert.Equal(t, "register_sensor", payload.Type)
	assert.Equal(t, "measurement", payload.Data.StateClass)
	assert.Equal(t, "cpu_usage", payload.Data.UniqueID)
}

func TestUpdateSensorDataDoesNotIncludeStateClass(t *testing.T) {
	var payload updateSensorRequestPayload
	client := NewAPI("http://example.com", "token", "device", true)
	client.Registration = Registration{WebhookID: "abc123"}
	client.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &payload))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{}")),
			Header:     make(http.Header),
		}, nil
	})

	err := client.UpdateSensorData(context.Background(), []UpdateSensorDataRequest{{
		Type:       "sensor",
		State:      42,
		Attributes: map[string]interface{}{"unit": "%"},
		UniqueID:   "cpu_usage",
		Icon:       "mdi:gauge",
	}})
	require.NoError(t, err)

	require.Len(t, payload.Data, 1)
	assert.Equal(t, "update_sensor_states", payload.Type)
	assert.Equal(t, "cpu_usage", payload.Data[0].UniqueID)
	assert.Equal(t, "mdi:gauge", payload.Data[0].Icon)
}
