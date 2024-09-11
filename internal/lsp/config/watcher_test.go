package config

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	configFilePath := tempDir + "/config.yaml"

	configFileContents := `---
foo: bar
`

	err := os.WriteFile(configFilePath, []byte(configFileContents), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	watcher := NewWatcher(&WatcherOpts{ErrorWriter: os.Stderr})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := watcher.Start(ctx)
		if err != nil {
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

	err = os.WriteFile(configFilePath, []byte(newConfigFileContents), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-watcher.Reload:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for config event")
	}

	err = os.Rename(configFilePath, configFilePath+".new")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-watcher.Drop:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for config drop event")
	}
}
