package sensor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"hacompanion/entity"
)

var reUptimeBootTimeDarwin = regexp.MustCompile(`sec = (\d+)`)

type Uptime struct{}

func NewUptime() *Uptime {
	return &Uptime{}
}

func (u Uptime) Run(ctx context.Context) (*entity.Payload, error) {
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "sysctl", "kern.boottime")
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return u.process(out.String())
}

func (u Uptime) process(output string) (*entity.Payload, error) {
	p := entity.NewPayload()
	match := reUptimeBootTimeDarwin.FindStringSubmatch(output)
	if len(match) < 2 {
		return nil, fmt.Errorf("failed to parse boot time from sysctl output: %s", output)
	}
	sec, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return nil, err
	}
	bootTime := time.Unix(sec, 0)
	uptime := time.Since(bootTime).Seconds()
	p.State = bootTime.Format(time.RFC3339)
	p.Attributes["uptime_seconds"] = strconv.FormatFloat(uptime, 'f', 2, 64)
	return p, nil
}
