package server

import (
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	ListenAddr string
}

func New(handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         "127.0.0.1:8080",
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		ListenAddr: "127.0.0.1:8080",
	}
}

func (s *Server) Start() error {
	if s.ListenAddr != "" {
		s.httpServer.Addr = s.ListenAddr
	}
	return s.httpServer.ListenAndServe()
}
