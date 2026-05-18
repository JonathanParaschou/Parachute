package server

import (
	"log"
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
			Addr:         "0.0.0.0:8080",
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		ListenAddr: "0.0.0.0:8080",
	}
}

func (s *Server) Start() error {
	if s.ListenAddr != "" {
		s.httpServer.Addr = s.ListenAddr
	}
	log.Printf("server listening on http://%s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &loggingResponseWriter{ResponseWriter: w}

		next.ServeHTTP(lw, r)

		status := lw.status
		if status == 0 {
			status = http.StatusOK
		}
		log.Printf("%s %s %d %dB %s remote=%s", r.Method, r.URL.RequestURI(), status, lw.bytes, time.Since(start).Round(time.Millisecond), r.RemoteAddr)
	})
}
