package main

import (
	"testing"

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
	output := &payload{
		State: 468.0,
		Attributes: map[string]interface{}{
			"mem_total": 15897.5,
			"mem_available": 15897.5,
			"swap_total": 16268,
			"swap_free": 15305,
		},
	}

	av := NewMemory()

	res, err := av.process(input)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
