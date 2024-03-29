package sensor

import (
	"testing"

	"hacompanion/entity"

	"github.com/stretchr/testify/require"
)

func TestMemory(t *testing.T) {
	input := `
		MemTotal:       16279032 kB
		MemFree:          479256 kB
		MemAvailable:    4469240 kB
		Buffers:         1003708 kB
		SwapTotal:      16658428 kB
		SwapFree:       15672316 kB
	`
	output := &entity.Payload{
		State: 468.02,
		Attributes: map[string]interface{}{
			"mem_total":     15897.49,
			"mem_available": 4364.49,
			"swap_total":    16267.99,
			"swap_free":     15304.99,
		},
	}

	av := NewMemory()

	res, err := av.process(input)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
