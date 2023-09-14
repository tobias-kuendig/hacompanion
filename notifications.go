package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"hacompanion/api"
	"hacompanion/util"
)

// NotificationServer listens for incoming notifications from Home Assistant.
type NotificationServer struct {
	registration api.Registration
	mux          *http.ServeMux
	address      string
	Server       *http.Server
}

func NewNotificationServer(registration api.Registration, address string) *NotificationServer {
	s := &NotificationServer{
		registration: registration,
		mux:          http.NewServeMux(),
		address:      address,
	}
	s.Server = &http.Server{
		Addr:    s.address,
		Handler: s.mux,
		// Set some reasonable timeouts, mostly to resolve linter
		// warnings about the potential for a slowloris DoS.
		ReadHeaderTimeout: time.Duration(10) * time.Second,
		ReadTimeout:       time.Duration(20) * time.Second,
		WriteTimeout:      time.Duration(20) * time.Second,
	}
	return s
}

func (s NotificationServer) Listen(ctx context.Context) {
	s.mux.HandleFunc("/notifications", func(w http.ResponseWriter, r *http.Request) {
		var notification Notification
		var req api.PushNotificationRequest
		w.Header().Set("Content-Type", "application/json")
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("received notification payload: %+v", req)
		if req.PushToken != s.registration.PushToken {
			log.Printf("notification push token is wrong: %s", req.PushToken)
			util.RespondError(w, "wrong token", http.StatusUnauthorized)
			return
		}
		err = notification.Send(r.Context(), req.Title, req.Message, req.Data)
		if err != nil {
			log.Printf("failed to send notification: %s", err)
			util.RespondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("notification sent successfully")
		w.WriteHeader(http.StatusCreated)
		util.RespondSuccess(w)
	})

	log.Printf("starting notification server on %s (with token %s)", s.address, s.registration.PushToken)

	if err := s.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to start notifications server: %s", err)
	}
}

// Notification is used to send notifications using native tools.
type Notification struct{}

func (n *Notification) Send(ctx context.Context, title, message string, data api.PushNotificationData) error {
	var args []string
	if data.Expire > 0 {
		args = append(args, "-t", strconv.Itoa(data.Expire))
	}
	if data.Urgency != "" {
		args = append(args, "-u", data.Urgency)
	}
	args = append(args, "-a", "'Home Assistant'")
	if title != "" {
		args = append(args, title)
	}
	args = append(args, message)
	log.Printf("comand is: notify-send %v", args)
	cmd := exec.CommandContext(ctx, "notify-send", args...)
	cmd.Env = []string{"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus"}
	if err := cmd.Run(); err != nil {
		log.Printf("Return :%+v", err)
		return err
	}
	return nil
}
