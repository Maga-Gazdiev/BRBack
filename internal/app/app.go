// Package app is the composition root: it wires infrastructure, services and HTTP.
package app

import (
	"context"
	"errors"
	"log/slog"
	"m96/internal/config"
	"m96/internal/handler/api"
	"m96/internal/infrastructure/jsonstore"
	"m96/internal/service/content"
	"net/http"
	"os"
	"time"
)

func Run(ctx context.Context, c config.Config) error {
	if e := os.MkdirAll(c.UploadDir, 0755); e != nil {
		return e
	}
	repo, e := jsonstore.New(c.DataFile)
	if e != nil {
		return e
	}
	h := api.New(content.New(repo), c.AdminToken, c.UploadDir)
	server := &http.Server{Addr: c.HTTPAddr, Handler: routes(h, c.UploadDir), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	done := make(chan error, 1)
	go func() { slog.Info("M96 API started", "address", c.HTTPAddr); done <- server.ListenAndServe() }()
	select {
	case e := <-done:
		if errors.Is(e, http.ErrServerClosed) {
			return nil
		}
		return e
	case <-ctx.Done():
		stop, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(stop)
	}
}
