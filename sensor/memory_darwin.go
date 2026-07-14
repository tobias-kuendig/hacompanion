package sensor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"hacompanion/entity"
	"hacompanion/util"
)

var reTopPhysMemDarwin = regexp.MustCompile(`(?m)^PhysMem:\s+([\d.]+)([KMG])\s+used.*?,\s+([\d.]+)([KMG])\s+unused`)

func parseTopMemoryMB(value, unit string) (float64, error) {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	switch strings.ToUpper(unit) {
	case "K":
		return num / 1024, nil
	case "M":
		return num, nil
	case "G":
		return num * 1024, nil
	default:
		return 0, fmt.Errorf("unsupported memory unit %s", unit)
	}
}

type Memory struct{}

func NewMemory() *Memory {
	return &Memory{}
}

func (m Memory) Run(ctx context.Context) (*entity.Payload, error) {
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "top", "-l", "1", "-n", "0")
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return m.process(out.String())
}

func (m Memory) process(output string) (*entity.Payload, error) {
	p := entity.NewPayload()

	match := reTopPhysMemDarwin.FindStringSubmatch(output)
	if len(match) < 5 {
		return nil, fmt.Errorf("failed to parse memory from top output: %s", output)
	}
	usedMB, err := parseTopMemoryMB(match[1], match[2])
	if err != nil {
		return nil, err
	}
	unusedMB, err := parseTopMemoryMB(match[3], match[4])
	if err != nil {
		return nil, err
	}

	usedMB = util.RoundToTwoDecimals(usedMB)
	totalMB := util.RoundToTwoDecimals(usedMB + unusedMB)
	p.State = usedMB
	p.Attributes["mem_total"] = totalMB

	return p, nil
}
