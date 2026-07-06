package sensor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"hacompanion/entity"
)

var (
	reAudioVolumeDarwin = regexp.MustCompile(`output volume:\s*(\d+)`)
	reAudioMutedDarwin  = regexp.MustCompile(`output muted:\s*(true|false)`)
)

type AudioVolume struct{}

func NewAudioVolume() *AudioVolume {
	return &AudioVolume{}
}

func (a AudioVolume) Run(ctx context.Context) (*entity.Payload, error) {
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "osascript", "-e", "get volume settings")
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return a.process(out.String())
}

func (a AudioVolume) process(output string) (*entity.Payload, error) {
	p := entity.NewPayload()

	volumeMatch := reAudioVolumeDarwin.FindStringSubmatch(output)
	if len(volumeMatch) < 2 {
		return nil, fmt.Errorf("failed to parse volume from osascript output: %s", output)
	}
	p.State = volumeMatch[1]

	mutedMatch := reAudioMutedDarwin.FindStringSubmatch(output)
	if len(mutedMatch) < 2 {
		return nil, fmt.Errorf("failed to parse muted state from osascript output: %s", output)
	}
	// match Linux behaviour: attribute is "muted", value is "on" when muted
	muted := "off"
	if mutedMatch[1] == "true" {
		muted = "on"
	}
	p.Attributes["muted"] = muted

	return p, nil
}
