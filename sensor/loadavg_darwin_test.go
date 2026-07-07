package sensor

import (
	"testing"

	"hacompanion/entity"

	"github.com/stretchr/testify/require"
)

func TestLoadAVG(t *testing.T) {
	sysctlOutput := `vm.loadavg: { 1.23 2.34 3.45 }`
	output := &entity.Payload{
		State: float64(1.23),
		Attributes: map[string]interface{}{
			"5m":  float64(2.34),
			"15m": float64(3.45),
		},
	}

	loadavg := NewLoadAVG()

	res, err := loadavg.process(sysctlOutput)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
