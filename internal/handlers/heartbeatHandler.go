package handlers

import (
	"net/http"
	"parachute/internal/services"
)

func Heartbeat(w http.ResponseWriter, r *http.Request) {
	heartbeatService := services.NewHeartbeatService()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(heartbeatService.Ping()))
}
