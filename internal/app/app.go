package app

import (
	"fmt"
	"net/http"
	"strings"

	"parachute/internal/config"
	"parachute/internal/handlers"
	"parachute/internal/server"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	router := http.NewServeMux()

	// routes
	router.HandleFunc("/heartbeat", handlers.Heartbeat)
	router.HandleFunc("/storage-metadata", handlers.StorageMetadata)
	router.HandleFunc("/storage-roots", handlers.StorageRoots)
	router.HandleFunc("/upload", handlers.Upload)

	// server instantiation
	srv := server.New(router)

	// If VPN is configured, also listen on VPN interface
	if cfg.VPN != nil {
		vpnIP := strings.Split(cfg.VPN.ServerIP, "/")[0]
		vpnAddr := fmt.Sprintf("%s:8080", vpnIP)
		fmt.Printf("Starting ParaChute server on http://%s (VPN)\n", vpnAddr)
		go func() {
			vpnSrv := server.New(router)
			vpnSrv.ListenAddr = vpnAddr
			if err := vpnSrv.Start(); err != nil {
				fmt.Printf("VPN server failed: %v\n", err)
			}
		}()
	}

	return srv.Start()
}
