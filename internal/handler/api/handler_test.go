package api_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"m96/internal/handler/api"
	"m96/internal/infrastructure/jsonstore"
	"m96/internal/model"
	"m96/internal/service/content"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const token = "test-token-at-least-32-characters-long"

func setup(t *testing.T) (*http.ServeMux, *content.Service, model.Content) {
	t.Helper()
	b, e := os.ReadFile("../../../data/content.json")
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "content.json")
	if e = os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
	r, e := jsonstore.New(path)
	if e != nil {
		t.Fatal(e)
	}
	s := content.New(r)
	c, e := s.Get(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	mux := http.NewServeMux()
	api.New(s, token, dir).Register(mux)
	return mux, s, c
}
func call(m http.Handler, method, path, key string, b []byte) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	if key != "" {
		r.Header.Set("Authorization", "Bearer "+key)
	}
	w := httptest.NewRecorder()
	m.ServeHTTP(w, r)
	return w
}
func TestAuthorizationAndValidation(t *testing.T) {
	mux, s, c := setup(t)
	b, _ := json.Marshal(c)
	for _, key := range []string{"", "wrong"} {
		if w := call(mux, "PUT", "/api/admin/content", key, b); w.Code != 401 {
			t.Fatalf("unauthorized = %d", w.Code)
		}
	}
	if w := call(mux, "GET", "/api/content", "", nil); w.Code != 200 || strings.Contains(w.Body.String(), token) {
		t.Fatal("public content failed or secret leaked")
	}
	c.Services[0].Price = -1
	b, _ = json.Marshal(c)
	if w := call(mux, "PUT", "/api/admin/content", token, b); w.Code != 422 {
		t.Fatalf("invalid price = %d", w.Code)
	}
	saved, _ := s.Get(context.Background())
	if saved.Revision != 1 {
		t.Fatal("invalid write modified content")
	}
}
func TestUpdateConflictAndUnsafeURLs(t *testing.T) {
	mux, _, c := setup(t)
	c.Brand = "Updated"
	b, _ := json.Marshal(c)
	if w := call(mux, "PUT", "/api/admin/content", token, b); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := call(mux, "PUT", "/api/admin/content", token, b); w.Code != 409 {
		t.Fatalf("stale update=%d", w.Code)
	}
	c.Revision++
	c.Settings.BookingURL = "https://yclients.com.evil.example/"
	b, _ = json.Marshal(c)
	if w := call(mux, "PUT", "/api/admin/content", token, b); w.Code != 422 {
		t.Fatal("unsafe booking URL accepted")
	}
	c.Settings.BookingURL = ""
	c.Hero.Image = "javascript:alert(1)"
	b, _ = json.Marshal(c)
	if w := call(mux, "PUT", "/api/admin/content", token, b); w.Code != 422 {
		t.Fatal("unsafe image URL accepted")
	}
}
func TestUploadAndMalformedBody(t *testing.T) {
	mux, _, _ := setup(t)
	if w := call(mux, "POST", "/api/admin/uploads", token, []byte("<svg onload='alert(1)'/>")); w.Code != 415 {
		t.Fatal("active image accepted")
	}
	b, e := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII=")
	if e != nil {
		t.Fatal(e)
	}
	w := call(mux, "POST", "/api/admin/uploads", token, b)
	if w.Code != 201 || !strings.Contains(w.Body.String(), "/uploads/") {
		t.Fatal(w.Body.String())
	}
	for _, b := range []string{`{"revision":1,"unexpected":true}`, `{} {}`} {
		if w := call(mux, "PUT", "/api/admin/content", token, []byte(b)); w.Code != 400 {
			t.Fatalf("malformed body=%d", w.Code)
		}
	}
}

func TestServerMetadataEscapesEditorText(t *testing.T) {
	mux, _, c := setup(t)
	c.Settings.Title = `M96 "</title><script>alert(1)</script>`
	c.Settings.SiteURL = "https://m96.example"
	b, _ := json.Marshal(c)
	if w := call(mux, "PUT", "/api/admin/content", token, b); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	w := call(mux, "GET", "/api/meta", "", nil)
	if w.Code != 200 || strings.Contains(w.Body.String(), "<script>") || !strings.Contains(w.Body.String(), "https://m96.example/images/hero.jpg") {
		t.Fatal(w.Body.String())
	}
	if w := call(mux, "GET", "/sitemap.xml", "", nil); w.Code != 200 || !strings.Contains(w.Body.String(), "https://m96.example") {
		t.Fatal(w.Body.String())
	}
}
