package sensor

import (
	"bufio"
	"bytes"
	"context"
	"hacompanion/entity"
	"log"
	"os/exec"
	"strings"
)

type Script struct {
	cfg entity.ScriptConfig
}

func NewScriptRunner(cfg entity.ScriptConfig) *Script {
	return &Script{
		cfg: cfg,
	}
}

func (s Script) Run(ctx context.Context) (*entity.Payload, error) {
	var err error
	var out bytes.Buffer

	// Call the custom script.
	cmd := exec.CommandContext(ctx, s.cfg.Path)
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		return nil, err
	}

	n := 0
	p := entity.NewPayload()
	sc := bufio.NewScanner(strings.NewReader(out.String()))
	for sc.Scan() {
		n++
		line := strings.TrimSpace(sc.Text())
		// Read in first line as the state.
		if n == 1 {
			if s.cfg.Type == "binary_sensor" {
				// Convert string to bool for a binary sensor.
				lineLower := strings.ToLower(line)
				strtobool := map[string]bool{"on": true, "true": true, "yes": true}
				p.State = strtobool[lineLower]
			} else {
				// No conversion needed for a regular sensor.
				p.State = line
			}
			continue
		}
		// Read in any additional lines as attributes.
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			log.Printf("ignoring custom script line with less than two parts: %s\n", line)
			continue
		}
		attrName := strings.TrimSpace(parts[0])
		attrValue := strings.TrimSpace(strings.Join(parts[1:], ":"))
		if attrName == "icon" {
			p.Icon = attrValue
		} else {
			p.Attributes[attrName] = attrValue
		}
	}
	return p, nil
}
