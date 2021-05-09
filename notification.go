package main

import (
	"context"
	"os/exec"
	"strconv"
)

type Notification struct{}

func (n *Notification) Send(ctx context.Context, title, message string, data PushNotificationData) error {
	var args []string
	if data.Expire > 0 {
		args = append(args, "-t", strconv.Itoa(data.Expire))
	}
	if data.Urgency != "" {
		args = append(args, "-u", data.Urgency)
	}
	args = append(args, "-a", "Home Assistant")
	if title != "" {
		args = append(args, title)
	}
	args = append(args, message)
	cmd := exec.CommandContext(ctx, "notify-send", args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
