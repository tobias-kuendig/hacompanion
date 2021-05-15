package main

import "hacompanion/entity"

// Config contains all values from the configuration file.
type Config struct {
	HomeAssistant homeassistantConfig            `toml:"homeassistant"`
	Companion     companionConfig                `toml:"companion"`
	Notifications notificationsConfig            `toml:"notifications"`
	Sensors       map[string]entity.SensorConfig `toml:"sensor"`
}

type homeassistantConfig struct {
	DeviceName string `toml:"device_name"`
	Token      string `toml:"token"`
	Host       string `toml:"host"`
}

type companionConfig struct {
	UpdateInterval   duration `toml:"update_interval"`
	RegistrationFile homePath `toml:"registration_file"`
}

type notificationsConfig struct {
	Listen  string `toml:"listen"`
	PushURL string `toml:"push_url"`
}
