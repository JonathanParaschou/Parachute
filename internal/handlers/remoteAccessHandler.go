package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"parachute/internal/services"
)

func RemoteAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	remoteAccessService := services.NewRemoteAccessService()
	status, err := remoteAccessService.Status()
	if err != nil {
		log.Println("Error reading remote access status:", err)
		http.Error(w, "Error reading remote access status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
