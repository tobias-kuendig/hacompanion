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
)

type Daemon struct {
	config Config
	cli    http.Client
}

func (d *Daemon) process(ctx context.Context, outputs chan Output) {
	for {
		select {
		case output := <-outputs:
			err := d.send(ctx, output)
			if err != nil {
				log.Printf("failed to send output: %s", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (d *Daemon) send(ctx context.Context, output Output) error {
	log.Printf("sending %+v", output)
	j, err := json.Marshal(output.payload)
	if err != nil {
		return err
	}
	log.Printf("payload %s", string(j))
	url := getAPIEndpoint(d.config.Host, output.integration.domain, d.config.Prefix, output.integration.name, output.payload.Name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+d.config.Token)
	resp, err := d.cli.Do(req)
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

func getAPIEndpoint(host, domain, prefix, service, payloadName string) string {
	if payloadName != "" {
		payloadName = "_" + payloadName
	}
	return fmt.Sprintf(
		"%s/api/states/%s.%s%s%s",
		strings.Trim(host, "/"),
		strings.TrimSpace(domain),
		strings.TrimSpace(prefix),
		strings.TrimSpace(service),
		strings.TrimSpace(payloadName),
	)
}
