package sensor

import (
	"testing"

	"hacompanion/entity"

	"github.com/stretchr/testify/require"
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
	output := &entity.Payload{
		State: "75",
		Attributes: map[string]interface{}{
			"muted": "off",
		},
	}

	av := NewAudioVolume()

	res, err := av.process(input)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
