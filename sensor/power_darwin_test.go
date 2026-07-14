package sensor

import (
	"testing"

	"hacompanion/entity"

	"github.com/stretchr/testify/require"
)

func TestPower_Battery(t *testing.T) {
	input := `Now drawing from 'Battery Power' -InternalBattery-0 (id=22413411) 74%; discharging; 4:00 remaining present: true`
	output := &entity.Payload{
		Icon:  "mdi:battery-70",
		State: "74",
		Attributes: map[string]interface{}{
			"ac_connected":   "off",
			"status":         "discharging",
			"time_remaining": "4:00",
		},
	}

	pwr := NewPower(entity.Meta{})

	res, err := pwr.process(input)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}

func TestPower_AC(t *testing.T) {
	input := `Now drawing from 'AC Power' -InternalBattery-0 (id=22413411) 83%; charging; 1:12 remaining present: true`
	output := &entity.Payload{
		Icon:  "mdi:battery-charging-80",
		State: "83",
		Attributes: map[string]interface{}{
			"ac_connected":   "on",
			"status":         "charging",
			"time_remaining": "1:12",
		},
	}

	pwr := NewPower(entity.Meta{})

	res, err := pwr.process(input)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
