package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"m96/internal/model"
	"m96/internal/service/content"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Handler struct {
	service *content.Service
	token   [32]byte
	uploads string
}

func New(s *content.Service, token, uploads string) *Handler {
	return &Handler{service: s, token: sha256.Sum256([]byte(token)), uploads: uploads}
}
func (h *Handler) Register(m *http.ServeMux) {
	m.HandleFunc("GET /api/content", h.get)
	m.HandleFunc("GET /api/meta", h.metadata)
	m.HandleFunc("GET /robots.txt", h.robots)
	m.HandleFunc("GET /sitemap.xml", h.sitemap)
	m.Handle("PUT /api/admin/content", h.authorize(http.HandlerFunc(h.update)))
	m.Handle("POST /api/admin/uploads", h.authorize(http.HandlerFunc(h.upload)))
	m.Handle("GET /api/admin/session", h.authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]bool{"ok": true}) })))
}
func (h *Handler) authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		auth := r.Header.Get("Authorization")
		provided := sha256.Sum256([]byte(strings.TrimPrefix(auth, "Bearer ")))
		if !strings.HasPrefix(auth, "Bearer ") || subtle.ConstantTimeCompare(provided[:], h.token[:]) != 1 {
			writeError(w, 401, "Неверный ключ доступа")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	c, e := h.service.Get(r.Context())
	if e != nil {
		h.fail(w, e)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	write(w, 200, c)
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		writeError(w, 415, "Ожидается application/json")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	var c model.Content
	if e := d.Decode(&c); e != nil {
		writeError(w, 400, "Некорректный JSON или размер больше 1 МБ")
		return
	}
	if e := d.Decode(new(any)); e != io.EOF {
		writeError(w, 400, "Ожидается один JSON объект")
		return
	}
	saved, e := h.service.Update(r.Context(), c)
	switch {
	case errors.Is(e, content.ErrConflict):
		writeError(w, 409, "Контент уже изменён. Скопируйте свои правки и перезагрузите страницу.")
	case errors.Is(e, content.ErrInvalid):
		writeError(w, 422, e.Error())
	case e != nil:
		h.fail(w, e)
	default:
		write(w, 200, saved)
	}
}
func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8*1024*1024)
	b, e := io.ReadAll(r.Body)
	if e != nil {
		writeError(w, 413, "Максимальный размер — 8 МБ")
		return
	}
	ext := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}[http.DetectContentType(b)]
	if ext == "" {
		writeError(w, 415, "Разрешены JPEG, PNG и WebP")
		return
	}
	id := make([]byte, 16)
	if _, e = rand.Read(id); e != nil {
		h.fail(w, e)
		return
	}
	name := hex.EncodeToString(id) + ext
	if e = os.WriteFile(filepath.Join(h.uploads, name), b, 0644); e != nil {
		h.fail(w, e)
		return
	}
	write(w, 201, map[string]string{"url": "/uploads/" + name})
}
func (h *Handler) fail(w http.ResponseWriter, e error) {
	slog.Error("request failed", "error", e)
	writeError(w, 500, "Не удалось сохранить или загрузить данные. Попробуйте ещё раз.")
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, message string) {
	write(w, status, map[string]string{"error": message})
}
