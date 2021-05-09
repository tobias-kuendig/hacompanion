package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
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
	parts := bytes.Fields(b)
	if len(parts) < 3 {
		return nil, fmt.Errorf("expected at least 3 values from /proc/loadavg, got only %d: %s", len(parts), string(b))
	}
	for index, load := range parts {
		switch index {
		case 0:
			p.State = string(load)
		case 1:
			p.Attributes["5m"] = string(load)
		case 2:
			p.Attributes["15m"] = string(load)
		default:
			break
		}
	}
	return p, nil
}
