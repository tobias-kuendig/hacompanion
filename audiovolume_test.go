package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAudioVolume(t *testing.T) {
	input := `
		Simple mixer control 'Master',0
		  Capabilities: pvolume pswitch pswitch-joined
		  Playback channels: Front Left - Front Right
		  Limits: Playback 0 - 65536
		  Mono:
		  Front Left: Playback 49151 [75%] [on]
		  Front Right: Playback 49151 [75%] [on]
	`
	output := &payloads{data: []payload{
		{
			Name:  "percentage",
			State: "75",
			Attributes: map[string]string{
				"unit_of_measurement": "%",
				"firendly_name": "Volume Level Percentage",
			},
		},
		{
			Name:  "muted",
			State: "off",
			Attributes: map[string]string{
				"firendly_name": "Volume Muted",
			},
		},
	}}

	av := NewAudioVolume()

	res, err := av.process(input)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
