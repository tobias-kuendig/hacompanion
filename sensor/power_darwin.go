package sensor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"hacompanion/entity"
)

var (
	rePowerPercent  = regexp.MustCompile(`\b(\d{1,3})%;`)
	rePowerStatus   = regexp.MustCompile(`;\s*(charging|discharging|charged|finishing charge|ac attached)\s*;`)
	rePowerTimeLeft = regexp.MustCompile(`(\d+:\d+) remaining`)
	rePowerSource   = regexp.MustCompile(`'([^']+)'`)
)

func NewPower(m entity.Meta) *Power {
	return &Power{}
}

func (pwr Power) Run(ctx context.Context) (*entity.Payload, error) {
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "pmset", "-g", "batt")
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return pwr.process(out.String())
}

func (pwr Power) process(output string) (*entity.Payload, error) {
	p := entity.NewPayload()

	percentMatch := rePowerPercent.FindStringSubmatch(output)
	if len(percentMatch) < 2 {
		return nil, fmt.Errorf("failed to parse battery percentage from pmset output: %s", output)
	}
	p.State = percentMatch[1]

	charging := false
	statusMatch := rePowerStatus.FindStringSubmatch(strings.ToLower(output))
	if len(statusMatch) >= 2 {
		status := statusMatch[1]
		p.Attributes["status"] = status
		charging = status == "charging" || status == "finishing charge"
	}

	sourceMatch := rePowerSource.FindStringSubmatch(output)
	if len(sourceMatch) >= 2 {
		if sourceMatch[1] == "AC Power" {
			p.Attributes["ac_connected"] = "on"
			charging = true
		} else {
			p.Attributes["ac_connected"] = "off"
		}
	}

	timeMatch := rePowerTimeLeft.FindStringSubmatch(output)
	if len(timeMatch) >= 2 {
		p.Attributes["time_remaining"] = timeMatch[1]
	}

	p.Icon = pwr.resolveIcon(p.State, charging)
	return p, nil
}
