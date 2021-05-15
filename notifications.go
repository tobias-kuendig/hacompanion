package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type NotificationServer struct {
	registration Registration
	mux          *http.ServeMux
	address      string
	server       *http.Server
}

func NewNotificationServer(registration Registration) *NotificationServer {
	s := &NotificationServer{
		registration: registration,
		mux:          http.NewServeMux(),
		address:      ":8080",
	}
	s.server = &http.Server{
		Addr:    s.address,
		Handler: s.mux,
	}
	return s
}

func (s NotificationServer) Listen(ctx context.Context) {
	s.mux.HandleFunc("/notification", func(w http.ResponseWriter, r *http.Request) {
		var notification Notification
		var req PushNotificationRequest
		w.Header().Set("Content-Type", "application/json")
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			respondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.PushToken != s.registration.PushToken {
			respondError(w, "wrong token", http.StatusUnauthorized)
			return
		}
		err = notification.Send(r.Context(), req.Title, req.Message, req.Data)
		if err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(201)
		respondSuccess(w)
	})

	err := s.server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
