package app

import (
	"m96/internal/handler/api"
	"net/http"
	"path/filepath"
	"regexp"
)

func routes(h *api.Handler, uploads string) http.Handler {
	m := http.NewServeMux()
	h.Register(m)
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	valid := regexp.MustCompile(`^[a-f0-9]{32}\.(jpg|png|webp)$`)
	m.HandleFunc("GET /uploads/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if !valid.MatchString(name) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.ServeFile(w, r, filepath.Join(uploads, name))
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		m.ServeHTTP(w, r)
	})
}
