package main

import (
	"hacompanion/entity"
	"hacompanion/sensor"
)

// sensorDefinitions is used to map the configuration to internal types.
var sensorDefinitions = map[string]func(m entity.Meta) entity.SensorDefinition{
	"cpu_temp": func(m entity.Meta) entity.SensorDefinition {
		unit := "C"
		if !m.GetBool("celsius") {
			unit = "F"
		}
		return entity.SensorDefinition{
			Type:        "sensor",
			Runner:      func(m entity.Meta) entity.Runner { return sensor.NewCPUTemp(m) },
			DeviceClass: "temperature",
			Icon:        "mdi:thermometer",
			Unit:        unit,
		}
	},
	"cpu_usage": func(m entity.Meta) entity.SensorDefinition {
		return entity.SensorDefinition{
			Type:   "sensor",
			Runner: func(m entity.Meta) entity.Runner { return sensor.NewCPUUsage() },
			Icon:   "mdi:gauge",
			Unit:   "%",
		}
	},
	"memory": func(m entity.Meta) entity.SensorDefinition {
		return entity.SensorDefinition{
			Type:   "sensor",
			Runner: func(meta entity.Meta) entity.Runner { return sensor.NewMemory() },
			Icon:   "mdi:memory",
		}
	},
	"power": func(m entity.Meta) entity.SensorDefinition {
		return entity.SensorDefinition{
			Type:        "sensor",
			Runner:      func(m entity.Meta) entity.Runner { return sensor.NewPower(m) },
			Icon:        "mdi:battery",
			DeviceClass: "battery",
			Unit:        "%",
		}
	},
	"uptime": func(m entity.Meta) entity.SensorDefinition {
		return entity.SensorDefinition{
			Runner:      func(meta entity.Meta) entity.Runner { return sensor.NewUptime() },
			Type:        "sensor",
			Icon:        "mdi:sort-clock-descending",
			DeviceClass: "timestamp",
		}
	},
	"load_avg": func(m entity.Meta) entity.SensorDefinition {
		return entity.SensorDefinition{
			Type:   "sensor",
			Runner: func(meta entity.Meta) entity.Runner { return sensor.NewLoadAVG() },
			Icon:   "mdi:gauge",
		}
	},
	"webcam": func(m entity.Meta) entity.SensorDefinition {
		return entity.SensorDefinition{
			Type:   "sensor",
			Runner: func(meta entity.Meta) entity.Runner { return sensor.NewWebCam() },
			Icon:   "mdi:webcam",
		}
	},
	"audio_volume": func(m entity.Meta) entity.SensorDefinition {
		return entity.SensorDefinition{
			Type:   "sensor",
			Runner: func(meta entity.Meta) entity.Runner { return sensor.NewAudioVolume() },
			Icon:   "mdi:volume-high",
			Unit:   "%",
		}
	},
	"online_check": func(m entity.Meta) entity.SensorDefinition {
		return entity.SensorDefinition{
			Type:   "binary_sensor",
			Runner: func(m entity.Meta) entity.Runner { return sensor.NewOnlineCheck(m) },
			Icon:   "mdi:shield-check-outline",
		}
	},
	"companion_running": func(m entity.Meta) entity.SensorDefinition {
		return entity.SensorDefinition{
			Type:   "binary_sensor",
			Runner: func(meta entity.Meta) entity.Runner { return NullRunner{} },
			Icon:   "mdi:heart-pulse",
		}
	},
}
