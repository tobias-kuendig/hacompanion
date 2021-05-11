package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

var reMemory = regexp.MustCompile(`(?mi)^\s?(?P<name>[^:]+):\s+(?P<value>\d+)`)

type Memory struct{}

func NewMemory() *Memory {
	return &Memory{}
}

func (m Memory) run(ctx context.Context) (*payload, error) {
	b, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	return m.process(string(b))
}

func (m Memory) process(output string) (*payload, error) {
	p := NewPayload()
	matches := reMemory.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		var err error
		var kb int
		if len(match) != 3 {
			continue
		}
		kb, err = strconv.Atoi(strings.TrimSpace(match[2]))
		if err != nil {
			continue
		}
		// convert kb to MB
		mb := roundToTwoDecimals(float64(kb) / 1024)
		switch strings.TrimSpace(match[1]) {
		case "MemFree":
			p.State = mb
		case "MemAvailable":
			fallthrough
		case "MemTotal":
			fallthrough
		case "SwapFree":
			fallthrough
		case "SwapTotal":
			p.Attributes[ToSnakeCase(match[1])] = mb
		}
	}
	if p.State == "" {
		return nil, fmt.Errorf("could not detrmine memory state based on /proc/meminfo: %s", output)
	}
	return p, nil
}
