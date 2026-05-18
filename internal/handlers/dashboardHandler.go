package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Dashboard(distDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		indexPath := filepath.Join(distDir, "index.html")
		requestPath := strings.TrimPrefix(filepath.Clean("/"+r.URL.Path), string(filepath.Separator))
		if requestPath == "." || requestPath == "" {
			http.ServeFile(w, r, indexPath)
			return
		}

		assetPath := filepath.Join(distDir, requestPath)
		info, err := os.Stat(assetPath)
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, assetPath)
			return
		}

		if _, err := os.Stat(indexPath); err != nil {
			http.Error(w, "dashboard build not found; run npm install && npm run build in web", http.StatusServiceUnavailable)
			return
		}
		http.ServeFile(w, r, indexPath)
	})
}
