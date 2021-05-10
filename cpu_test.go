package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCPUUsage(t *testing.T) {
	inputs := []string{
		`
		cpu     1000 54353    2000    5000 292242 0 388396 0 0 0
		cpu0 1676694 12408 1035902 9278196 73034 0 143838 0 0 0
		cpu1 1679176 13614 1057959 9239394 74962 0 96468 0 0 0
		cpu2 1671376 14174 1049842 9271404 71242 0 82559 0 0 0
		cpu3 1688265 14156 1057602 9151504 73002 0 65530 0 0 0
		`, `
		cpu     2000 54353    4000   10000 292242 0 388396 0 0 0
		cpu0 1676694 12408 1035902 9278215 73034 0 143838 0 0 0
		cpu1 1679176 13614 1057959 9239413 74962 0 96468 0 0 0
		cpu2 1671376 14174 1049843 9271423 71242 0 82559 0 0 0
		cpu3 1688265 14156 1057603 9151522 73002 0 65530 0 0 0
		`,
	}
	// u1=1000+2000     t1=1000+2000+5000
	// u2=2000+4000     t2=2000+4000+10000
	// u=(u2-u1) * 100 / (t2-t1) = % usage
	output := &payload{
		State: "37.5",
		Attributes: map[string]interface{}{
//			"core_0": "34.0",
//			"core_1": "35.0",
//			"core_2": "36.0",
//			"core_3": "37.0",
//			"core_4": "38.0",
//			"core_5": "39.0",
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
	output := &payload{
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

	c := NewCPUTemp(Meta{"celsius": true})

	res, err := c.process(input)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
