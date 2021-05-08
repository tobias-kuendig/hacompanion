package main

import (
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"strings"
)

type CPUTemp struct{}

func NewCPUTemp() *CPUTemp {
	return &CPUTemp{}
}

func (c CPUTemp) run(ctx context.Context) (payload, error) {
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "cat", "/sys/class/thermal/thermal_zone4/temp")
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return payload{}, err
	}
	milli, err := strconv.Atoi(strings.TrimSpace(out.String()))
	if err != nil {
		return payload{}, err
	}
	cent := milli / 1000
	return payload{
		State: strconv.Itoa(cent),
		Attributes: map[string]string{
			"friendly_name":       "Linux CPU Temp",
			"unit_of_measurement": "C",
			"device_class":        "temperature",
		},
	}, nil
}
