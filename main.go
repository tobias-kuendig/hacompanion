package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

var host = "192.168.1.77:8123"
var token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiIzZGQyNGFhMWZiOWY0NjhjYWU2YjBmMDYwMGVkOGU1ZiIsImlhdCI6MTYyMDIzODA3MSwiZXhwIjoxOTM1NTk4MDcxfQ.acv3qSoz9IdLVN6oJofTafMPovqGri-L1Efrk3Q255w"
var cli http.Client

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("failed to start: %s", err)
	}
}

type IntegrationConfig struct {
	Enabled bool
	Name    string
	Meta    map[string]interface{}
}
type payload struct {
	State      string            `json:"state,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}
type runner interface {
	run(ctx context.Context) (payload, error)
}

type Integration struct {
	ticker *time.Ticker
	domain string
	name   string
	runner runner
}

func (i Integration) String() string {
	return fmt.Sprintf("%s.%s", i.domain, i.name)
}

type Output struct {
	payload     payload
	integration Integration
}

type payloads struct {
	data []payload
}

func Payloads() *payloads {
	return &payloads{
		data: make([]payload),
	}
}

func (p *Payloads) Multiple () *Payloads {
	return p
}

func (p *Payloads) New (payload payload) payload {
	p.data = append(p.data, payload)
	return payload
}
func (p *Payloads) Add (payload payload) payload {
	p.data = append(p.data, payload)
	return payload
}

func run(ctx context.Context) error {
	var config struct {
		Integrations map[string]IntegrationConfig `toml:"integration"`
	}
	b, err := ioutil.ReadFile("hadaemon.toml")
	if err != nil {
		return err
	}
	if _, err := toml.Decode(string(b), &config); err != nil {
		return err
	}
	results := make(chan Output, 5)
	go process(ctx, results)

	runners := map[string]struct {
		runner func() runner
		domain string
	}{
		"cpu_temp": {
			runner: func() runner { return NewCPUTemp() },
			domain: "sensor",
		},
		"webcam": {
			runner: func() runner { return NewWebCam() },
			domain: "sensor",
		},
	}

	var integrations []Integration
	for key, integrationConfig := range config.Integrations {
		if !integrationConfig.Enabled {
			continue
		}
		if _, ok := runners[key]; !ok {
			return fmt.Errorf("unknown integration %s in config", key)
		}
		integrations = append(integrations, Integration{
			ticker: time.NewTicker(8 * time.Second),
			name:   integrationConfig.Name,
			domain: runners[key].domain,
			runner: runners[key].runner(),
		})
	}

	var wg sync.WaitGroup
	for _, integration := range integrations {
		wg.Add(1)
		go integration.start(ctx, &wg, results)
	}

	wg.Wait()
	return nil
}

func (i Integration) start(ctx context.Context, wg *sync.WaitGroup, outputs chan Output) {
	defer wg.Done()

	fn := func() {
		value, err := i.runner.run(ctx)
		if err != nil {
			log.Printf("failed to run integration %s: %s", i, err)
			return
		}
		log.Printf("received payload: %+v", value)
		outputs <- Output{integration: i, payload: value}
	}

	fn()

	for {
		select {
		case <-i.ticker.C:
			fn()
		case <-ctx.Done():
			return
		}
	}
}

func send(ctx context.Context, output Output) error {
	log.Printf("sending %+v", output)
	j, err := json.Marshal(output.payload)
	if err != nil {
		return err
	}
	log.Printf("payload %s", string(j))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url(host, output.integration.domain, output.integration.name), bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode > 201 {
		return errors.New(fmt.Sprintf("received invalid status code %d (%s)", resp.StatusCode, body))
	}
	log.Printf("received %s", string(body))
	return nil
}

func process(ctx context.Context, outputs chan Output) {
	for {
		select {
		case output := <-outputs:
			err := send(ctx, output)
			if err != nil {
				log.Printf("failed to send output: %s", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func url(host, domain, service string) string {
	return fmt.Sprintf(
		"http://%s/api/states/%s.%s",
		strings.Trim(host, "/"),
		strings.TrimSpace(domain),
		strings.TrimSpace(service),
	)
}
