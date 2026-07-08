package sensor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	"hacompanion/entity"
	"hacompanion/util"
)

var reCPUUsageDarwin = regexp.MustCompile(`CPU usage:\s+([\d.]+)%\s+user,\s+([\d.]+)%\s+sys`)

type CPUTemp struct{}

func NewCPUTemp(_ entity.Meta) *CPUTemp {
	return &CPUTemp{}
}

func (c CPUTemp) Run(_ context.Context) (*entity.Payload, error) {
	p := entity.NewPayload()
	p.State = "unavailable"
	return p, nil
}

type CPUUsage struct{}

func NewCPUUsage() *CPUUsage {
	return &CPUUsage{}
}

func (c CPUUsage) Run(ctx context.Context) (*entity.Payload, error) {
	var out bytes.Buffer
	// The first top sample is an uptime average, so request two and use the last one.
	cmd := exec.CommandContext(ctx, "top", "-l", "2", "-n", "0")
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return c.process(out.String())
}

func (c CPUUsage) process(output string) (*entity.Payload, error) {
	p := entity.NewPayload()
	matches := reCPUUsageDarwin.FindAllStringSubmatch(output, -1)
	if len(matches) < 2 {
		return nil, fmt.Errorf("expected at least 2 cpu usage samples from top output, got %d: %s", len(matches), output)
	}
	last := matches[len(matches)-1]
	user, err := strconv.ParseFloat(last[1], 64)
	if err != nil {
		return nil, err
	}
	sys, err := strconv.ParseFloat(last[2], 64)
	if err != nil {
		return nil, err
	}
	p.State = util.RoundToTwoDecimals(user + sys)
	return p, nil
}
