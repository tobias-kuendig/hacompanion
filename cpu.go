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

func (c CPUTemp) run(ctx context.Context) (*payload, error) {
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

func (c CPUTemp) process(output string) (*payload, error) {
	p := NewPayload()
	matches := reCPUTemp.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) < 3 {
			return nil, fmt.Errorf("invalid output form lm-sensors received: %s", output)
		}
		if strings.EqualFold(match[1], "Package id") {
			p.State = match[2]
		} else {
			p.Attributes[ToSnakeCase(match[1])] = match[2]
		}
	}
	if p.State == "" {
		return nil, fmt.Errorf("failed to parse cpu temperature state out of lm-sensors output: %s", output)
	}
	return p, nil
}
