package main

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"
)

var reAudioVolume = regexp.MustCompile(`(?m)Playback \d+ \[(?P<volume>\d{1,3})%\] \[(?P<state>on|off)\]`)

type AudioVolume struct{}

func NewAudioVolume() *AudioVolume {
	return &AudioVolume{}
}

func (a AudioVolume) run(ctx context.Context) (*payloads, error) {
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
	for _, flag := range flags {
		args = append(args, flag)
	}
	cmd := exec.CommandContext(ctx, "amixer", args...)
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func (a AudioVolume) process(output string) (*payloads, error) {
	matches := reAudioVolume.FindStringSubmatch(output)
	result := make(map[string]string)
	for i, name := range reAudioVolume.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = matches[i]
		}
	}
	p := MultiplePayloads()
	if volume, ok := result["volume"]; ok {
		p.Add(payload{
			Name:       "percentage",
			State:      volume,
			Attributes: map[string]string{"unit_of_measurement": "%", "friendly_name": "Volume Level Percentage"},
		})
	}
	if muted, ok := result["state"]; ok {
		// switch the muted state: amixer returns "on" for an enabled device and "off" for a muted device.
		// since we care about the "muted" state, we have to always send the opposite value.
		if muted == "off" {
			muted = "on"
		} else {
			muted = "off"
		}
		p.Add(payload{
			Name:  "muted",
			State: muted,
			Attributes: map[string]string{"friendly_name": "Volume Muted"},
		})
	}
	return p, nil
}
