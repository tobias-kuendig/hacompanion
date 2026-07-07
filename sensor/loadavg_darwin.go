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

var reLoadAVGDarwin = regexp.MustCompile(`\{\s*([\d.]+)\s+([\d.]+)\s+([\d.]+)\s*\}`)

type LoadAVG struct{}

func NewLoadAVG() *LoadAVG {
	return &LoadAVG{}
}

func (w LoadAVG) Run(ctx context.Context) (*entity.Payload, error) {
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "sysctl", "vm.loadavg")
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return w.process(out.String())
}

func (w LoadAVG) process(output string) (*entity.Payload, error) {
	p := entity.NewPayload()
	match := reLoadAVGDarwin.FindStringSubmatch(output)
	if len(match) < 4 {
		return nil, fmt.Errorf("failed to parse load averages from sysctl output: %s", output)
	}
	for index, load := range match[1:4] {
		float, err := strconv.ParseFloat(load, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse load value %s: %w", load, err)
		}
		float = util.RoundToTwoDecimals(float)
		switch index {
		case 0:
			p.State = float
		case 1:
			p.Attributes["5m"] = float
		case 2:
			p.Attributes["15m"] = float
		}
	}
	return p, nil
}
