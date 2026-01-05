package app

import (
	"net/http"

	"parachute/internal/handlers"
	"parachute/internal/server"
)

func Run() error {
	router := http.NewServeMux()

	// routes
	router.HandleFunc("/heartbeat", handlers.Heartbeat)
	router.HandleFunc("/storage-metadata", handlers.StorageMetadata)

	// server instantiation
	srv := server.New(router)

	return srv.Start()
}
