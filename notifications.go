package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
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
	tags         map[string][]string
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
	s.tags = make(map[string][]string)
	return s
}

func (s NotificationServer) Listen(_ context.Context) {
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
		var id string
		if req.Message == "clear_notification" {
			idlist := s.tags[req.Data.Tag]
			if len(idlist) > 0 {
				var anysuccess bool = false
				for _, id = range idlist {
					err = notification.Clear(r.Context(), id)
					if err == nil {
						anysuccess = true
					} else {
						log.Printf("failed to clear notification: %s", err)
					}
				}
				if !anysuccess {
					log.Println("clear_notification failed")
					util.RespondError(w, "clear_notification failed", http.StatusInternalServerError)
					return
				}
				delete(s.tags, req.Data.Tag)
			} else {
				log.Printf("No notification found with tag %s", req.Data.Tag)
			}
			log.Println("notification sent successfully")
		} else {
			err, id = notification.Send(r.Context(), req.Title, req.Message, req.Data)
			if err != nil {
				log.Printf("failed to send notification: %s", err)
				util.RespondError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if req.Data.Tag != "" {
				s.tags[req.Data.Tag] = append([]string{id}, s.tags[req.Data.Tag]...)
			}
			log.Println("notification cleared successfully")
		}
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

func (n *Notification) Send(ctx context.Context, title, message string, data api.PushNotificationData) (error, string) {
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
	args = append(args, "-h", "string:desktop-entry:hacompanion")
	args = append(args, message)
	args = append(args, "-p")
	log.Printf("comand is: notify-send %v", args)
	cmd := exec.CommandContext(ctx, "notify-send", args...)
	cmd.Env = []string{"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus"}
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Return :%+v", err)
		return err, ""
	}
	id := strings.TrimSpace(string(out[:]))
	return nil, id
}

func (n *Notification) Clear(ctx context.Context, id string) error {
	var args []string
	args = append(args, "call", "--session", "--dest", "org.freedesktop.Notifications", "--object-path", "/org/freedesktop/Notifications", "--method",  "org.freedesktop.Notifications.CloseNotification")
	args = append(args, id)
	log.Printf("comand is: gdbus %v", args)
	cmd := exec.CommandContext(ctx, "gdbus", args...)
	cmd.Env = []string{"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus"}
	if err := cmd.Run(); err != nil {
		log.Printf("Return :%+v", err)
		return err
	}
	return nil
}
