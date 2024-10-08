package config

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/styrainc/regal/internal/lsp/log"
)

func TestWatcher(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	configFilePath := tempDir + "/config.yaml"

	configFileContents := `---
foo: bar
`

	if err := os.WriteFile(configFilePath, []byte(configFileContents), 0o600); err != nil {
		t.Fatal(err)
	}

	watcher := NewWatcher(&WatcherOpts{LogFunc: func(l log.Level, s string, a ...any) {
		t.Logf(l.String()+": "+s, a...)
	}})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := watcher.Start(ctx); err != nil {
			t.Errorf("failed to start watcher: %v", err)
		}
	}()

	watcher.Watch(configFilePath)

	select {
	case <-watcher.Reload:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for initial config event")
	}

	newConfigFileContents := `---
foo: baz
`

	if err := os.WriteFile(configFilePath, []byte(newConfigFileContents), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case <-watcher.Reload:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for config event")
	}

	if err := os.Rename(configFilePath, configFilePath+".new"); err != nil {
		t.Fatal(err)
	}

	select {
	case <-watcher.Drop:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for config drop event")
	}
}
