package sensor

import (
	"testing"

	"hacompanion/entity"

	"github.com/stretchr/testify/require"
)

func TestCPUUsage(t *testing.T) {
	topOutput := `CPU usage: 12.50% user, 7.25% sys, 80.25% idle
CPU usage: 4.20% user, 3.91% sys, 91.89% idle`

	output := &entity.Payload{
		State:      8.11,
		Attributes: map[string]interface{}{},
	}

	c := NewCPUUsage()

	res, err := c.process(topOutput)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
