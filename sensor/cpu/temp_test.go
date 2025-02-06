package cpu

import (
	"testing"

	"hacompanion/entity"

	"github.com/stretchr/testify/require"
)

func TestCPUTemp_Intel(t *testing.T) {
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
			"temp_1": "37.0",
		},
	}

	c := NewCPUTemp(entity.Meta{"celsius": true})

	res, err := c.process(input)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}

func TestCPUTemp_AMD(t *testing.T) {
	input := `
        k10temp-pci-00c3
		Adapter: PCI adapter
		Tctl:         +39.6°C	

		acpitz-acpi-0
		Adapter: ACPI interface
		temp1:        +27.8°C  (crit = +119.0°C)
	`
	output := &entity.Payload{
		State: "39.6",
		Attributes: map[string]interface{}{
			"temp_1": "27.8",
		},
	}

	c := NewCPUTemp(entity.Meta{"celsius": true})

	res, err := c.process(input)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
