package jsonstore_test

import (
	"context"
	"encoding/json"
	"errors"
	"m96/internal/infrastructure/jsonstore"
	"m96/internal/model"
	"m96/internal/service/content"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func seed(t *testing.T) (string, model.Content) {
	t.Helper()
	b, e := os.ReadFile("../../../data/content.json")
	if e != nil {
		t.Fatal(e)
	}
	var c model.Content
	if e = json.Unmarshal(b, &c); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "content.json")
	if e = os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
	return path, c
}
func TestConcurrentEditorsAndPersistence(t *testing.T) {
	path, c := seed(t)
	store, e := jsonstore.New(path)
	if e != nil {
		t.Fatal(e)
	}
	var successes, conflicts atomic.Int32
	var wg sync.WaitGroup
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := store.Save(context.Background(), c)
			switch {
			case e == nil:
				successes.Add(1)
			case errors.Is(e, content.ErrConflict):
				conflicts.Add(1)
			default:
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 || conflicts.Load() != 11 {
		t.Fatalf("success=%d conflicts=%d", successes.Load(), conflicts.Load())
	}
	reopened, e := jsonstore.New(path)
	if e != nil {
		t.Fatal(e)
	}
	saved, e := reopened.Get(context.Background())
	if e != nil || saved.Revision != c.Revision+1 {
		t.Fatalf("persisted revision=%d err=%v", saved.Revision, e)
	}
}
func TestCancelledSaveDoesNotModifyContent(t *testing.T) {
	path, c := seed(t)
	store, e := jsonstore.New(path)
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e = store.Save(ctx, c); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	saved, e := store.Get(context.Background())
	if e != nil || saved.Revision != c.Revision {
		t.Fatal("cancelled write changed data")
	}
}
