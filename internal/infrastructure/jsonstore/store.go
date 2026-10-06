// Package jsonstore provides single-process durable content storage.
package jsonstore

import (
	"context"
	"encoding/json"
	"fmt"
	"m96/internal/model"
	"m96/internal/service/content"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	path string
}

func New(path string) (*Store, error) {
	s := &Store{path: path}
	c, err := s.Get(context.Background())
	if err != nil {
		return nil, err
	}
	if err = content.Validate(c); err != nil {
		return nil, fmt.Errorf("validate initial content: %w", err)
	}
	return s, nil
}
func (s *Store) read() (model.Content, error) {
	var c model.Content
	b, e := os.ReadFile(s.path)
	if e != nil {
		return c, e
	}
	e = json.Unmarshal(b, &c)
	return c, e
}
func (s *Store) Get(ctx context.Context) (model.Content, error) {
	if e := ctx.Err(); e != nil {
		return model.Content{}, e
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.read()
}
func (s *Store) Save(ctx context.Context, c model.Content) (model.Content, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := ctx.Err(); e != nil {
		return model.Content{}, e
	}
	old, e := s.read()
	if e != nil {
		return model.Content{}, e
	}
	if c.Revision != old.Revision {
		return model.Content{}, content.ErrConflict
	}
	c.Revision++
	b, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		return model.Content{}, e
	}
	f, e := os.CreateTemp(filepath.Dir(s.path), ".content-*")
	if e != nil {
		return model.Content{}, e
	}
	defer os.Remove(f.Name())
	if _, e = f.Write(b); e != nil {
		f.Close()
		return model.Content{}, e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return model.Content{}, e
	}
	if e = f.Close(); e != nil {
		return model.Content{}, e
	}
	if e = os.Rename(f.Name(), s.path); e != nil {
		return model.Content{}, e
	}
	return c, nil
}
