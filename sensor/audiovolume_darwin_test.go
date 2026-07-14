package sensor

import (
	"testing"

	"hacompanion/entity"

	"github.com/stretchr/testify/require"
)

func TestAudioVolume(t *testing.T) {
	input := `output volume:19, input volume:50, alert volume:100, output muted:false`
	output := &entity.Payload{
		State: "19",
		Attributes: map[string]interface{}{
			"muted": "off",
		},
	}

	av := NewAudioVolume()

	res, err := av.process(input)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
