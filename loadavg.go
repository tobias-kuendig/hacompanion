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

func (w LoadAVG) run(ctx context.Context) (*payloads, error) {
	b, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	p := MultiplePayloads()
	attributes := []struct {
		Name         string
		FriendlyName string
	}{
		{Name: "1m", FriendlyName: "Load Average (1 min)"},
		{Name: "5m", FriendlyName: "Load Average (5 min)"},
		{Name: "15m", FriendlyName: "Load Average (15 min)"},
	}
	parts := bytes.Fields(b)
	if len(parts) < 3 {
		return nil, fmt.Errorf("expected at least 3 values from /proc/loadavg, got only %d: %s", len(parts), string(b))
	}
	for index, load := range parts {
		attrs := attributes[index]
		p.Add(payload{
			Name:  attrs.Name,
			State: string(load),
			Attributes: map[string]string{
				"friendly_name": attrs.FriendlyName,
			},
		})
		if index >= 2 {
			break
		}
	}
	return p, nil
}
