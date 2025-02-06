package cpu

import (
	"bytes"
	"context"
	"fmt"
	"hacompanion/entity"
	"hacompanion/util"
	"os/exec"
	"regexp"
	"strings"
)

var (
	reCPUTemp = regexp.MustCompile(`(?m)(temp1|Package id|Core \d|CPU|Tctl)[\s\d]*:\s+.?([\d\.]+)°`)
	// This is currently unused.
	//reCPUTemp2 = regexp.MustCompile(`(?mi)^\s?(?P<name>[^:]+):\s+(?P<value>\d+)`)
	reCPUUsage = regexp.MustCompile(`(?m)^\s*cpu(\d+)?.*`)
)

type Temp struct {
	UseCelsius bool
}

func NewCPUTemp(m entity.Meta) *Temp {
	c := &Temp{}
	if m.GetBool("celsius") {
		c.UseCelsius = true
	}
	return c
}

func (c Temp) Run(ctx context.Context) (*entity.Payload, error) {
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

func (c Temp) process(output string) (*entity.Payload, error) {
	p := entity.NewPayload()
	matches := reCPUTemp.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) < 3 {
			return nil, fmt.Errorf("invalid output form lm-sensors received: %s", output)
		}
		if strings.EqualFold(match[1], "Package id") || strings.EqualFold(match[1], "CPU") || strings.EqualFold(match[1], "Tctl") {
			p.State = match[2]
		} else {
			p.Attributes[util.ToSnakeCase(match[1])] = match[2]
		}
	}
	if p.State == "" {
		return nil, fmt.Errorf("failed to parse cpu temperature state out of lm-sensors output: %s", output)
	}
	return p, nil
}
