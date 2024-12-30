package sensor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"hacompanion/entity"
)

type OnlineCheck struct {
	mode   string
	target string
	client http.Client
}

func NewOnlineCheck(m entity.Meta) *OnlineCheck {
	o := OnlineCheck{
		mode: "ping",
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
	if mode := m.GetString("mode"); mode != "" {
		o.mode = mode
	}
	if host := m.GetString("target"); host != "" {
		o.target = host
	}
	return &o
}

func (o OnlineCheck) Run(ctx context.Context) (*entity.Payload, error) {
	if o.target == "" {
		return nil, fmt.Errorf("online check requires target to be specified")
	}
	switch o.mode {
	case "http":
		return o.checkHTTP(ctx)
	case "ping":
		return o.checkPing(ctx)
	default:
		return nil, fmt.Errorf("unknown mode for online check: %s", o.mode)
	}
}

func (o OnlineCheck) checkHTTP(ctx context.Context) (*entity.Payload, error) {
	p := entity.NewPayload()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "HomeAssistant-Companion/Online-Check")

	resp, err := o.client.Do(req)
	if err != nil {
		p.State = false
		p.Attributes["err"] = err.Error()
		return p, nil
	}

	defer resp.Body.Close()

	p.State = true
	p.Attributes["status"] = resp.Status

	return p, nil
}

func (o OnlineCheck) checkPing(ctx context.Context) (*entity.Payload, error) {
	p := entity.NewPayload()
	//nolint:gosec
	cmd := exec.CommandContext(ctx, "ping", "-c 2", "-w 4", o.target)
	err := cmd.Run()
	if err != nil {
		p.State = false
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			p.Attributes["err"] = fmt.Sprintf("could not reach %s", o.target)
		}
		return p, nil
	}

	p.State = true
	return p, nil
}
