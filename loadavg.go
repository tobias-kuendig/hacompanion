package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
)

type LoadAVG struct{}

func NewLoadAVG() *LoadAVG {
	return &LoadAVG{}
}

func (w LoadAVG) run(ctx context.Context) (*payload, error) {
	b, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	p := NewPayload()
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
			// Round to two decimal places.
			float = math.Floor(float * 100) / 100
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
