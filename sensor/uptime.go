package sensor

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"hacompanion/entity"
)

type Uptime struct{}

func NewUptime() *Uptime {
	return &Uptime{}
}

func (u Uptime) Run(ctx context.Context) (*entity.Payload, error) {
	b, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return nil, err
	}
	return u.process(string(b))
}

func (u Uptime) process(output string) (*entity.Payload, error) {
	p := entity.NewPayload()
	parts := strings.Fields(output)
	if len(parts) < 2 {
		return nil, fmt.Errorf("expected at least two values from /proc/uptime: %s", output)
	}
	seconds, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse seconds from /proc/uptime (%s): %s", output, err)
	}

	p.State = time.Now().Add(-time.Second * time.Duration(seconds)).Format(time.RFC3339)
	p.Attributes["uptime_seconds"] = parts[0]

	return p, nil
}
