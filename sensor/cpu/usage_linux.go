package cpu

import (
	"context"
	"fmt"
	"hacompanion/entity"
	"hacompanion/util"
	"os"
	"strconv"
	"strings"
	"time"
)

type Usage struct{}

func NewCPUUsage() *Usage {
	return &Usage{}
}

func (c Usage) Run(ctx context.Context) (*entity.Payload, error) {
	var outputs []string
	measurements := 2
	for i := 0; i < measurements; i++ {
		b, err := os.ReadFile("/proc/stat")
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, string(b))
		// Don't sleep if this is the last iteration.
		if i < measurements-1 {
			time.Sleep(1 * time.Second)
		}
	}
	return c.process(outputs)
}

func (c Usage) process(outputs []string) (*entity.Payload, error) {
	p := entity.NewPayload()
	type stat struct {
		usage float64
		total float64
	}
	// Parse out the usage deltas from the stats output.
	stats := map[string][]stat{}
	for i, output := range outputs {
		// Return a measurement for a single core.
		matches := reCPUUsage.FindAllStringSubmatch(output, -1)
		for _, submatch := range matches {
			match := strings.TrimSpace(submatch[0])
			var cpu string
			if len(submatch) > 1 {
				cpu = strings.TrimSpace(submatch[1])
			}
			// Fetch the relevant values and convert them to floats.
			fields := strings.Fields(match)
			user, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				return nil, err
			}
			system, err := strconv.ParseFloat(fields[3], 64)
			if err != nil {
				return nil, err
			}
			idle, err := strconv.ParseFloat(fields[4], 64)
			if err != nil {
				return nil, err
			}
			// Calculate the effective usage as well as the available total.
			if stats[cpu] == nil {
				stats[cpu] = make([]stat, 2)
			}
			stats[cpu][i] = stat{
				usage: user + system,
				total: user + system + idle,
			}
		}
	}
	// Calculate the usage per core as a percentage.
	for cpu, value := range stats {
		u := value[1].usage - value[0].usage
		t := value[1].total - value[0].total
		if t > 0 {
			percent := util.RoundToTwoDecimals(u * 100 / t)
			if cpu == "" {
				p.State = percent
			} else {
				p.Attributes[fmt.Sprintf("core_%s", cpu)] = percent
			}
		}
	}
	return p, nil
}
