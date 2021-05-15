package sensor

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"hacompanion/entity"
	"hacompanion/util"
)

var (
	reCPUTemp  = regexp.MustCompile(`(?m)(Package id|Core \d)[\s\d]*:\s+.?([\d\.]+)°`)
	reCPUUsage = regexp.MustCompile(`(?m)^\s*cpu(\d+)?.*`)
)

type CPUTemp struct {
	UseCelsius bool
}

func NewCPUTemp(m entity.Meta) *CPUTemp {
	c := &CPUTemp{}
	if m.GetBool("celsius") {
		c.UseCelsius = true
	}
	return c
}

func (c CPUTemp) Run(ctx context.Context) (*entity.Payload, error) {
	var out bytes.Buffer
	var args []string
	if !c.UseCelsius {
		args = append(args, "--fahrenheit")
	}
	cmd := exec.CommandContext(ctx, "sensors", args...)
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return c.process(out.String())
}

func (c CPUTemp) process(output string) (*entity.Payload, error) {
	p := entity.NewPayload()
	matches := reCPUTemp.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) < 3 {
			return nil, fmt.Errorf("invalid output form lm-sensors received: %s", output)
		}
		if strings.EqualFold(match[1], "Package id") {
			p.State = match[2]
		} else {
			p.Attributes[util.ToSnakeCase(match[1])] = match[2]
		}
	}
	if p.State == "" {
		return nil, fmt.Errorf("failed to parse cpu temperature state out of lm-sensors output: %s", output)
	}
	return p, nil
}

type CPUUsage struct{}

func NewCPUUsage() *CPUUsage {
	return &CPUUsage{}
}

func (c CPUUsage) Run(ctx context.Context) (*entity.Payload, error) {
	var outputs []string
	measurements := 2
	for i := 0; i < measurements; i++ {
		b, err := ioutil.ReadFile("/proc/stat")
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

func (c CPUUsage) process(outputs []string) (*entity.Payload, error) {
	p := entity.NewPayload()
	type stat struct {
		usage float64
		total float64
	}
	// Parse the usage deltas out form the stats output.
	stats := map[string][]stat{}
	for i, output := range outputs {
		// Returns a single cpu core measurement
		matches := reCPUUsage.FindAllStringSubmatch(output, -1)
		for _, submatch := range matches {
			match := strings.TrimSpace(submatch[0])
			var cpu string
			if len(submatch) > 1 {
				cpu = strings.TrimSpace(submatch[1])
			}
			// Fetch the relevant values, convert them to floats.
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
	// Calculate the percentage values per core.
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
