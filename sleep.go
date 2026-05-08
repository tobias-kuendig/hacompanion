//go:build !windows
// +build !windows

package main

import (
	"log"
	"syscall"

	"github.com/godbus/dbus/v5"
)

func handleSleep(companion *Companion) {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Println("D-bus is unavailable. Not handling login inhibitor")
		return
	}
	defer conn.Close()

	var fd int
	obj := conn.Object(loginDestination, loginPath)

	err = obj.
		Call(loginMethodInhibit, 0, inhibitorMethodSleep, AppID, inhibitorMessage, inhibitorModeDelay).
		Store(&fd)

	if err != nil {
		log.Fatalf("error storing file descriptor: %v", err)
		return
	}

	err = conn.AddMatchSignal(
		dbus.WithMatchInterface(loginInterface),
		dbus.WithMatchObjectPath(loginPath),
		dbus.WithMatchMember("PrepareForSleep"),
	)

	if err != nil {
		log.Fatalf("error adding match signal: %v", err)
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)

	for sig := range c {
		isSleeping, ok := sig.Body[0].(bool)
		if !ok {
			continue
		}

		if isSleeping {
			companion.UpdateCompanionSensorData(false)
			err = syscall.Close(fd)
			if err != nil {
				log.Fatalf("error closing file descriptor: %v", err)
				return
			}
		} else {
			companion.UpdateCompanionSensorData(true)

			err = obj.
				Call(loginMethodInhibit, 0, inhibitorMethodSleep, AppID, inhibitorMessage, inhibitorModeDelay).
				Store(&fd)
			if err != nil {
				log.Fatalf("error storing file descriptor: %v", err)
				return
			}
		}
	}
}
