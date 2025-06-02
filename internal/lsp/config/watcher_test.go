package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/internal/testutil"
)

func TestWatcher(t *testing.T) {
	t.Parallel()

	tempDir := testutil.TempDirectoryOf(t, map[string]string{
		"config.yaml": `---
foo: bar
`,
	})

	watcher := NewWatcher(&WatcherOpts{LogFunc: func(l log.Level, s string, a ...any) {
		t.Logf(l.String()+": "+s, a...)
	}})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go func() {
		if err := watcher.Start(ctx); err != nil {
			t.Errorf("failed to start watcher: %v", err)
		}
	}()

	configFilePath := filepath.Join(tempDir, "config.yaml")

	watcher.Watch(configFilePath)

	select {
	case <-watcher.Reload:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for initial config event")
	}

	newConfigFileContents := `---
foo: baz
`
	testutil.MustWriteFile(t, configFilePath, []byte(newConfigFileContents))

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
