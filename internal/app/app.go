package app

import (
	"fmt"
	"net/http"

	"parachute/internal/config"
	"parachute/internal/handlers"
	"parachute/internal/server"
)

func Run() error {
	_, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	router := http.NewServeMux()

	// routes
	router.HandleFunc("/heartbeat", handlers.Heartbeat)
	router.HandleFunc("/storage-metadata", handlers.StorageMetadata)
	router.HandleFunc("/storage-roots", handlers.StorageRoots)
	router.HandleFunc("/upload", handlers.Upload)
	router.HandleFunc("/uploads", handlers.Uploads)
	router.HandleFunc("/remote-access", handlers.RemoteAccess)
	router.Handle("/", handlers.Dashboard("web/dist"))

	// server instantiation
	srv := server.New(server.LoggingMiddleware(router))

	return srv.Start()
}
