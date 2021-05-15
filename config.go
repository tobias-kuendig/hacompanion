package main

import "hacompanion/entity"

// Config contains all values from the configuration file.
type Config struct {
	HomeAssistant   homeassistantConfig            `toml:"homeassistant"`
	CompanionConfig companionConfig                `toml:"companion"`
	Sensors         map[string]entity.SensorConfig `toml:"sensor"`
}

type homeassistantConfig struct {
	DeviceName string `toml:"device_name"`
	Token      string `toml:"token"`
	Host       string `toml:"host"`
}

type companionConfig struct {
	UpdateInterval     duration `toml:"update_interval"`
	NotificationServer string   `toml:"notification_server"`
	RegistrationFile   homePath `toml:"registration_file"`
}
