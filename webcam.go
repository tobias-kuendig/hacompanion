package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type WebCam struct{}

func NewWebCam() *WebCam {
	return &WebCam{}
}

func (w WebCam) run(ctx context.Context) (payload, error) {
	var err error
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "lsmod")
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		return payload{}, err
	}
	var procCount string
	prefix := []byte("uvcvideo")
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		if !bytes.HasPrefix(scanner.Bytes(), prefix) {
			continue
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			return payload{}, fmt.Errorf("expected three values for lsmod uvcvideo entry, got %s", scanner.Text())
		}
		procCount = fields[2]
		break
	}
	if procCount == "" {
		return payload{}, errors.New("did no find uvcvideo in lsmod output, failed to determine webcam usage")
	}
	return payload{
		State: procCount,
		Attributes: map[string]string{
			"friendly_name": "Webcam process count",
		},
	}, nil
}
