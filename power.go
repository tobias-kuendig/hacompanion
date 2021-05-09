package main

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
)

type Power struct {
	Battery string
}

func NewPower(m Meta) *Power {
	c := &Power{Battery: "BAT0"}
	if b := m.GetString("battery"); b != "" {
		c.Battery = b
	}
	return c
}

func (pwr Power) run(ctx context.Context) (*payload, error) {
	dir := fmt.Sprintf("/sys/class/power_supply/%s", pwr.Battery)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read battery status from %s: %s", dir, err)
	}
	realPath, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to eval symlink %s: %s", dir, err)
	}
	p := NewPayload()
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
			p.Attributes["present"] = stringToOnOff(pwr.optimisticRead(filepath.Join(realPath, d.Name())))
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
	if exists, _ := fileExists(acLink); exists {
		if realPath, err := filepath.EvalSymlinks(acLink); err == nil {
			acInfo := filepath.Join(realPath, "online")
			if exists, _ := fileExists(acLink); exists {
				spew.Dump("READING", pwr.optimisticRead(acInfo))
				p.Attributes["ac_connected"] = stringToOnOff(pwr.optimisticRead(acInfo))
			}
		}
	}
	return p, err
}

func (pwr Power) optimisticRead(file string) string {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(b)
}
