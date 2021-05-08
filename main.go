package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

func main() {
	var config Config
	b, err := ioutil.ReadFile("hadaemon.toml")
	if err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}
	if _, err = toml.Decode(string(b), &config); err != nil {
		log.Fatalf("failed to parse config file: %s", err)
	}
	if err = run(context.Background(), config); err != nil {
		log.Fatalf("failed to start: %s", err)
	}
}

type Config struct {
	Prefix       string                       `toml:"prefix"`
	Token        string                       `toml:"token"`
	Host         string                       `toml:"host"`
	Integrations map[string]IntegrationConfig `toml:"integration"`
}

type IntegrationConfig struct {
	Enabled bool
	Name    string
	Meta    map[string]interface{}
}

type runner interface {
	run(ctx context.Context) (*payloads, error)
}

type Output struct {
	payload     payload
	integration Integration
}

type Meta map[string]interface{}

func (m Meta) GetBool(key string) bool {
	if v, ok := m[key]; ok {
		if v == true {
			return true
		}
		return false
	}
	return false
}

var runners = map[string]struct {
	runner func(Meta) runner
	domain string
}{
	"cpu_temp": {
		runner: func(m Meta) runner { return NewCPUTemp(m) },
		domain: "sensor",
	},
	"load_avg": {
		runner: func(Meta) runner { return NewLoadAVG() },
		domain: "sensor",
	},
	"webcam": {
		runner: func(Meta) runner { return NewWebCam() },
		domain: "sensor",
	},
	"audio_volume": {
		runner: func(Meta) runner { return NewAudioVolume() },
		domain: "sensor",
	},
}

func run(ctx context.Context, config Config) error {
	rand.Seed(time.Now().UnixNano())

	daemon := Daemon{
		config: config,
		cli:    http.Client{Timeout: 5 * time.Second},
	}

	integrations, err := buildIntegrations(config)
	if err != nil {
		return err
	}

	results := make(chan Output, 5)
	go daemon.process(ctx, results)

	var wg sync.WaitGroup
	for _, integration := range integrations {
		wg.Add(1)
		go integration.start(ctx, &wg, results)
	}

	wg.Wait()
	return nil
}

func buildIntegrations(config Config) ([]Integration, error) {
	var integrations []Integration
	for key, integrationConfig := range config.Integrations {
		if !integrationConfig.Enabled {
			continue
		}
		if _, ok := runners[key]; !ok {
			return nil, fmt.Errorf("unknown integration %s in config", key)
		}
		integrations = append(integrations, Integration{
			ticker: time.NewTicker(8 * time.Second),
			name:   integrationConfig.Name,
			domain: runners[key].domain,
			runner: runners[key].runner(integrationConfig.Meta),
		})
	}
	return integrations, nil
}

func getSmear() int {
	min := -8
	max := 8
	return rand.Intn(max-min+1) + min
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

func (i Integration) start(ctx context.Context, wg *sync.WaitGroup, outputs chan Output) {
	defer wg.Done()

	fn := func() {
		time.Sleep(time.Duration(getSmear()) * time.Second)
		values, err := i.runner.run(ctx)
		if err != nil {
			log.Printf("failed to run integration %s: %s", i, err)
			return
		}
		for _, value := range values.data {
			log.Printf("received payload: %+v", value)
			outputs <- Output{integration: i, payload: value}
		}
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
