package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
)

var reCPUTemp = regexp.MustCompile(`(?m)(Package id|Core \d)[\s\d]*:\s+.?([\d\.]+)°`)
var reCPUUsage = regexp.MustCompile(`(?m)^\s*cpu.*`)

type CPUTemp struct {
	UseCelsius bool
}

func NewCPUTemp(m Meta) *CPUTemp {
	c := &CPUTemp{}
	if m.GetBool("celsius") == true {
		c.UseCelsius = true
	}
	return c
}

func (c CPUTemp) run(ctx context.Context) (*payload, error) {
	var out bytes.Buffer
	var args []string
	if !c.UseCelsius {
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


type CPUUsage struct { }

func NewCPUUsage() *CPUUsage {
	return &CPUUsage{}
}

func (c CPUUsage) run(ctx context.Context) (*payload, error) {
	var outputs []string
	measurements := 2
	for i := 0; i < measurements; i++ {
		b, err := ioutil.ReadFile("/proc/stat")
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, string(b))
		// Don't sleep if this is the last iteration.
		if i < measurements - 1 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	return c.process(outputs)
}

func (c CPUUsage) process(outputs []string) (*payload, error) {
	p := NewPayload()
	for _, output := range outputs {
		match := reCPUUsage.FindString(output)
		match = strings.TrimSpace(match)
		fields := strings.Fields(match)
		spew.Dump(fields)
	}
	return p, nil
}
