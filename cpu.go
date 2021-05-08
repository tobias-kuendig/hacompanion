package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var reCPUTemp = regexp.MustCompile(`(?m)(Package id|Core \d)[\s\d]*:\s+.?([\d\.]+)°`)

type CPUTemp struct {
	UseCelcius bool
}

func NewCPUTemp(m Meta) *CPUTemp {
	c := &CPUTemp{}
	if m.GetBool("celcius") == true {
		c.UseCelcius = true
	}
	return c
}

func (c CPUTemp) run(ctx context.Context) (*payloads, error) {
	var out bytes.Buffer
	var args []string
	if !c.UseCelcius {
		args = append(args, "--fahrenheit")
	}
	cmd := exec.CommandContext(ctx, "sensors", args...)
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return c.process(out.String())
}

func (c CPUTemp) process(output string) (*payloads, error) {
	unit := "C"
	if !c.UseCelcius {
		unit = "F"
	}
	p := payload{}
	matches := reCPUTemp.FindAllStringSubmatch(output, -1)
	attrs := map[string]string{
		"unit_of_measurement": unit,
		"friendly_name":       "CPU Temperature",
		"device_class":        "temperature",
	}
	for _, match := range matches {
		if len(match) < 3 {
			return nil, fmt.Errorf("invalid output form lm-sensors received: %s", output)
		}
		if strings.EqualFold(match[1], "Package id") {
			p.State = match[2]
		} else {
			attrs[ToSnakeCase(match[1])] = match[2]
		}
	}
	if p.State == "" {
		return nil, fmt.Errorf("failed to parse cpu temperature state out of lm-sensors output: %s", output)
	}
	p.Attributes = attrs
	return SinglePayload(p), nil
}
