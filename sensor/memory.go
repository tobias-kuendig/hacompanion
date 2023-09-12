package sensor

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"hacompanion/entity"
	"hacompanion/util"
)

var reMemory = regexp.MustCompile(`(?mi)^\s?(?P<name>[^:]+):\s+(?P<value>\d+)`)

type Memory struct{}

func NewMemory() *Memory {
	return &Memory{}
}

func (m Memory) Run(ctx context.Context) (*entity.Payload, error) {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	return m.process(string(b))
}

func (m Memory) process(output string) (*entity.Payload, error) {
	p := entity.NewPayload()
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
		// Convert kb to MB.
		mb := util.RoundToTwoDecimals(float64(kb) / 1024)
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
			p.Attributes[util.ToSnakeCase(match[1])] = mb
		}
	}
	if p.State == "" {
		return nil, fmt.Errorf("could not determine memory state based on /proc/meminfo: %s", output)
	}
	return p, nil
}
