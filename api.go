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
	"time"
)

type RegisterDeviceRequest struct {
	DeviceID           string  `json:"device_id"`
	AppID              string  `json:"app_id"`
	AppName            string  `json:"app_name"`
	AppVersion         string  `json:"app_version"`
	DeviceName         string  `json:"device_name"`
	Manufacturer       string  `json:"manufacturer"`
	Model              string  `json:"model"`
	OsName             string  `json:"os_name"`
	OsVersion          string  `json:"os_version"`
	SupportsEncryption bool    `json:"supports_encryption"`
	AppData            AppData `json:"app_data"`
}

type AppData struct {
	PushToken string `json:"push_token"`
	PushURL   string `json:"push_url"`
}

type RegisterSensorRequest struct {
	Attributes        map[string]string `json:"attributes"`
	DeviceClass       string            `json:"device_class,omitempty"`
	Icon              string            `json:"icon"`
	Name              string            `json:"name"`
	State             string            `json:"state,omitempty"`
	Type              string            `json:"type"`
	UniqueId          string            `json:"unique_id"`
	UnitOfMeasurement string            `json:"unit_of_measurement"`
}

type registerSensorRequestPayload struct {
	Data RegisterSensorRequest `json:"data"`
	Type string                `json:"type"`
}

type UpdateSensorDataRequest struct {
	Attributes map[string]string `json:"attributes"`
	Icon       string            `json:"icon"`
	State      string            `json:"state"`
	Type       string            `json:"type"`
	UniqueId   string            `json:"unique_id"`
}

type updateSensorRequestPayload struct {
	Data []UpdateSensorDataRequest `json:"data"`
	Type string                    `json:"type"`
}

type Registration struct {
	CloudhookURL string `json:"cloudhook_url"`
	RemoteUIURL  string `json:"remote_ui_url"`
	Secret       string `json:"secret"`
	WebhookID    string `json:"webhook_id"`
	PushToken    string `json:"push_token"`
}

type PushNotificationRequest struct {
	Message          string `json:"message"`
	Title            string `json:"title"`
	PushToken        string `json:"push_token"`
	RegistrationInfo struct {
		AppId      string `json:"app_id"`
		AppVersion string `json:"app_version"`
		OsVersion  string `json:"os_version"`
	} `json:"registration_info"`
	Data PushNotificationData `json:"data"`
}

type PushNotificationData struct {
	Key     string `json:"key"`
	Urgency string `json:"urgency"`
	Expire  int    `json:"expire"`
}

func (api *API) URL(skipCloud bool) string {
	var url string
	if api.registration.CloudhookURL != "" && !skipCloud {
		url = api.registration.CloudhookURL
	} else if api.registration.RemoteUIURL != "" {
		url = fmt.Sprintf("%s/api/webhook/%s", api.registration.RemoteUIURL, api.registration.WebhookID)
	} else {
		url = fmt.Sprintf("%s/api/webhook/%s", api.Host, api.registration.WebhookID)
	}
	return url
}

func (r Registration) JSON() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type API struct {
	Host         string
	Token        string
	client       http.Client
	registration Registration
}

func NewAPI(host, token string) *API {
	return &API{
		Host:   host,
		Token:  token,
		client: http.Client{Timeout: 5 * time.Second},
	}
}

func (api *API) sendRequest(ctx context.Context, url string, payload []byte) ([]byte, error) {
	log.Printf("sending to %s: %+v", url, string(payload))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+api.Token)
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf("received invalid status code %d (%s)", resp.StatusCode, body))
	}
	log.Printf("received %s", string(body))
	return body, nil
}

func (api *API) RegisterDevice(ctx context.Context, request RegisterDeviceRequest) (Registration, error) {
	url := fmt.Sprintf("%s/api/mobile_app/registrations", strings.Trim(api.Host, "/"))
	var response Registration
	j, err := json.Marshal(request)
	if err != nil {
		return response, err
	}
	body, err := api.sendRequest(ctx, url, j)
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(body, &response)
	return response, err
}

func (api *API) RegisterSensor(ctx context.Context, data RegisterSensorRequest) error {
	if data.Attributes == nil {
		data.Attributes = make(map[string]string)
	}
	req := registerSensorRequestPayload{
		Data: data,
		Type: "register_sensor",
	}
	j, err := json.Marshal(req)
	if err != nil {
		return err
	}
	_, err = api.sendRequest(ctx, api.URL(false), j)
	if err != nil {
		_, err = api.sendRequest(ctx, api.URL(true), j)
	}
	return err
}

func (api *API) UpdateSensorData(ctx context.Context, data []UpdateSensorDataRequest) error {
	for key := range data {
		if data[key].Attributes == nil {
			data[key].Attributes = make(map[string]string)
		}
	}
	req := updateSensorRequestPayload{
		Data: data,
		Type: "update_sensor_states",
	}
	j, err := json.Marshal(req)
	if err != nil {
		return err
	}
	_, err = api.sendRequest(ctx, api.URL(false), j)
	if err != nil {
		_, err = api.sendRequest(ctx, api.URL(true), j)
	}
	return err
}

// RegisterSensors register's a slice of sensors in Home Assistant.
func (api *API) RegisterSensors(ctx context.Context, sensors []Sensor) error {
	for _, sensor := range sensors {
		err := api.RegisterSensor(ctx, RegisterSensorRequest{
			Type:              "sensor",
			DeviceClass:       sensor.deviceClass,
			Icon:              sensor.icon,
			Name:              sensor.name,
			UniqueId:          sensor.uniqueID,
			UnitOfMeasurement: sensor.unit,
		})
		if err != nil {
			return fmt.Errorf("failed to register sensor %s: %w", sensor.uniqueID, err)
		}
	}
	return nil
}
