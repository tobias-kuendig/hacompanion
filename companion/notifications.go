package companion

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strconv"

	"hacompanion/api"
	"hacompanion/util"
)

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

// NotificationServer listens for incoming notifications from Home Assistant.
type NotificationServer struct {
	registration api.Registration
	mux          *http.ServeMux
	address      string
	Server       *http.Server
}

func NewNotificationServer(registration api.Registration) *NotificationServer {
	s := &NotificationServer{
		registration: registration,
		mux:          http.NewServeMux(),
		address:      ":8080",
	}
	s.Server = &http.Server{
		Addr:    s.address,
		Handler: s.mux,
	}
	return s
}

func (s NotificationServer) Listen(ctx context.Context) {
	s.mux.HandleFunc("/notification", func(w http.ResponseWriter, r *http.Request) {
		var notification Notification
		var req api.PushNotificationRequest
		w.Header().Set("Content-Type", "application/json")
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.PushToken != s.registration.PushToken {
			util.RespondError(w, "wrong token", http.StatusUnauthorized)
			return
		}
		err = notification.Send(r.Context(), req.Title, req.Message, req.Data)
		if err != nil {
			util.RespondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(201)
		util.RespondSuccess(w)
	})

	err := s.Server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
