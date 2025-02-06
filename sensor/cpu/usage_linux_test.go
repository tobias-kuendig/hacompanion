package cpu

import (
	"testing"

	"hacompanion/entity"

	"github.com/stretchr/testify/require"
)

func TestCPUUsage(t *testing.T) {
	inputs := []string{
		`
		cpu     1000 54353    2000    5000 292242 0 388396 0 0 0
		cpu0    3000 12408    4000   10000 73034 0 143838 0 0 0
		cpu1    1000 13614    1000    2000 74962 0 96468 0 0 0
		`, `
		cpu     2000 54353    4000   10000 292242 0 388396 0 0 0
		cpu0    4000 12408    8000   20000 73034 0 143838 0 0 0
		cpu1    2000 13614    2000    5000 74962 0 96468 0 0 0
		`,
	}
	// u1=1000+2000     t1=1000+2000+5000
	// u2=2000+4000     t2=2000+4000+10000
	// u=(u2-u1) * 100 / (t2-t1) = % usage
	output := &entity.Payload{
		State: 37.5,
		Attributes: map[string]interface{}{
			"core_0": 33.33,
			"core_1": float64(40),
		},
	}

	c := NewCPUUsage()

	res, err := c.process(inputs)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
