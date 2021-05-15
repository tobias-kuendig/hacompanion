package sensor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"hacompanion/entity"
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

func TestCPUTemp(t *testing.T) {
	input := `
		coretemp-isa-0000
		Adapter: ISA adapter
		Package id 0:  +99.0°C  (high = +82.0°C, crit = +100.0°C)
		Core 0:        +34.0°C  (high = +82.0°C, crit = +100.0°C)
		Core 1:        +35.0°C  (high = +82.0°C, crit = +100.0°C)
		Core 2:        +36.0°C  (high = +82.0°C, crit = +100.0°C)
		Core 3:        +37.0°C  (high = +82.0°C, crit = +100.0°C)
		Core 4:        +38.0°C  (high = +82.0°C, crit = +100.0°C)
		Core 5:        +39.0°C  (high = +82.0°C, crit = +100.0°C)

		acpitz-acpi-0
		Adapter: ACPI interface
		temp1:        +27.8°C  (crit = +119.0°C)

		pch_cannonlake-virtual-0
		Adapter: Virtual device
		temp1:        +37.0°C  
	`
	output := &entity.Payload{
		State: "99.0",
		Attributes: map[string]interface{}{
			"core_0": "34.0",
			"core_1": "35.0",
			"core_2": "36.0",
			"core_3": "37.0",
			"core_4": "38.0",
			"core_5": "39.0",
		},
	}

	c := NewCPUTemp(entity.Meta{"celsius": true})

	res, err := c.process(input)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
