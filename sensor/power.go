package sensor

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"hacompanion/entity"
	"hacompanion/util"
)

type Power struct {
	Battery string
}

func NewPower(m entity.Meta) *Power {
	c := &Power{Battery: "BAT0"}
	if b := m.GetString("battery"); b != "" {
		c.Battery = b
	}
	return c
}

func (pwr Power) Run(ctx context.Context) (*entity.Payload, error) {
	dir := fmt.Sprintf("/sys/class/power_supply/%s", pwr.Battery)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read battery status from %s: %w", dir, err)
	}
	realPath, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to eval symlink %s: %w", dir, err)
	}
	p := entity.NewPayload()
	err = filepath.WalkDir(realPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch d.Name() {
		case "capacity":
			p.State = pwr.optimisticRead(filepath.Join(realPath, d.Name()))
		case "capacity_level":
			p.Attributes["level"] = pwr.optimisticRead(filepath.Join(realPath, d.Name()))
		case "present":
			p.Attributes["battery_present"] = util.StringToOnOff(pwr.optimisticRead(filepath.Join(realPath, d.Name())))
		case "status":
			p.Attributes["status"] = pwr.optimisticRead(filepath.Join(realPath, d.Name()))
		case "voltage_now":
			p.Attributes["voltage_now"] = pwr.optimisticRead(filepath.Join(realPath, d.Name()))
		case "voltage_min_design":
			p.Attributes["voltage_min_design"] = pwr.optimisticRead(filepath.Join(realPath, d.Name()))
		case "charge_now":
			p.Attributes["charge_now"] = pwr.optimisticRead(filepath.Join(realPath, d.Name()))
		case "charge_full":
			p.Attributes["charge_full"] = pwr.optimisticRead(filepath.Join(realPath, d.Name()))
		}
		return nil
	})
	// Check if a power cable is attached.
	acLink := "/sys/class/power_supply/AC"
	if exists, _ := util.FileExists(acLink); exists {
		if realPath, fileErr := filepath.EvalSymlinks(acLink); fileErr == nil {
			acInfo := filepath.Join(realPath, "online")
			if exists, _ := util.FileExists(acLink); exists {
				p.Attributes["ac_connected"] = util.StringToOnOff(pwr.optimisticRead(acInfo))
			}
		}
	}
	p.Icon = pwr.resolveIcon(p.State)
	return p, err
}

func (pwr Power) optimisticRead(file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(b)
}

func (pwr Power) resolveIcon(state any) string {
	num, err := strconv.Atoi(strings.TrimSpace(state.(string)))
	if err != nil {
		return "mdi:battery-unknown"
	}

	switch {
	case num >= 90:
		return "mdi:battery"
	case num >= 80:
		return "mdi:battery-80"
	case num >= 70:
		return "mdi:battery-70"
	case num >= 60:
		return "mdi:battery-60"
	case num >= 50:
		return "mdi:battery-50"
	case num >= 40:
		return "mdi:battery-40"
	case num >= 30:
		return "mdi:battery-30"
	case num >= 20:
		return "mdi:battery-20"
	default:
		return "mdi:battery-alert"
	}
}
