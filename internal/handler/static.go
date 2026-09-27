package handler

import (
	"net/http"
	"path/filepath"
)

// serveWebFile serves a file from the web/ directory (cwd = repo root when using go run).
func serveWebFile(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("web", name))
	}
}
