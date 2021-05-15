package sensor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"hacompanion/entity"
)

var reAudioVolume = regexp.MustCompile(`(?m)Playback \d+ \[(?P<volume>\d{1,3})%\]\s?(?:\[.+\])?\s?\[(?P<state>on|off)\]`)

type AudioVolume struct{}

func NewAudioVolume() *AudioVolume {
	return &AudioVolume{}
}

func (a AudioVolume) Run(ctx context.Context) (*entity.Payload, error) {
	var output string
	var err error
	output, err = a.getOutput(ctx)
	if err == nil {
		return a.process(output)
	}
	output, err = a.getOutput(ctx, "-D", "pulse")
	if err == nil {
		return a.process(output)
	}
	return nil, err
}

func (a AudioVolume) getOutput(ctx context.Context, flags ...string) (string, error) {
	var out bytes.Buffer
	args := []string{"sget", "Master"}
	args = append(args, flags...)
	cmd := exec.CommandContext(ctx, "amixer", args...)
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func (a AudioVolume) process(output string) (*entity.Payload, error) {
	p := entity.NewPayload()
	matches := reAudioVolume.FindStringSubmatch(output)
	result := make(map[string]string)
	names := reAudioVolume.SubexpNames()
	if len(names) != len(matches) {
		return nil, fmt.Errorf("failed to parse amixer output, regex did not return expected matches")
	}
	for i, name := range names {
		if i != 0 && name != "" {
			result[name] = matches[i]
		}
	}
	if volume, ok := result["volume"]; ok {
		p.State = volume
	}
	if muted, ok := result["state"]; ok {
		// switch the muted state: amixer returns "on" for an enabled device and "off" for a muted device.
		// since we care about the "muted" state, we have to always send the opposite value.
		if muted == "off" {
			muted = "on"
		} else {
			muted = "off"
		}
		p.Attributes["muted"] = muted
	}
	return p, nil
}
