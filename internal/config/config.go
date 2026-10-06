package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct{ HTTPAddr, DataFile, UploadDir, AdminToken string }

func Load() (Config, error) {
	c := Config{HTTPAddr: value("HTTP_ADDR", "127.0.0.1:8080"), DataFile: value("DATA_FILE", "data/content.json"), UploadDir: value("UPLOAD_DIR", "data/uploads"), AdminToken: os.Getenv("ADMIN_TOKEN")}
	if len(c.AdminToken) < 32 {
		return c, fmt.Errorf("ADMIN_TOKEN must contain at least 32 characters")
	}
	var e error
	c.DataFile, e = filepath.Abs(c.DataFile)
	if e != nil {
		return c, e
	}
	c.UploadDir, e = filepath.Abs(c.UploadDir)
	return c, e
}
func value(k, d string) string {
	if s := os.Getenv(k); s != "" {
		return s
	}
	return d
}
