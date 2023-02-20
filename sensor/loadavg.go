package sensor

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"hacompanion/entity"
	"hacompanion/util"
)

type LoadAVG struct{}

func NewLoadAVG() *LoadAVG {
	return &LoadAVG{}
}

func (w LoadAVG) Run(ctx context.Context) (*entity.Payload, error) {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	p := entity.NewPayload()
	parts := strings.Fields(string(b))
	if len(parts) < 3 {
		return nil, fmt.Errorf("expected at least 3 values from /proc/loadavg, got only %d: %s", len(parts), string(b))
	}
	for index, load := range parts {
		var float float64
		if index <= 2 {
			float, err = strconv.ParseFloat(load, 32)
			if err != nil {
				return nil, fmt.Errorf("failed to parse loadavg %s: %w", load, err)
			}
			float = util.RoundToTwoDecimals(float)
		}
		switch index {
		case 0:
			p.State = float
		case 1:
			p.Attributes["5m"] = float
		case 2:
			p.Attributes["15m"] = float
		default:
			break
		}
	}
	return p, nil
}
