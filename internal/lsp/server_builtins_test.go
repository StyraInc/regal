package lsp

import (
	"testing"

	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/pkg/config"
)

// https://github.com/StyraInc/regal/issues/679
func TestProcessBuiltinUpdateExitsOnMissingFile(t *testing.T) {
	t.Parallel()

	ls := NewLanguageServer(t.Context(), &LanguageServerOptions{Logger: log.NewLogger(log.LevelDebug, t.Output())})
	ls.loadedConfig = &config.Config{}

	if err := ls.processHoverContentUpdate(t.Context(), "file://missing.rego"); err != nil {
		t.Fatal(err)
	}

	if l := len(ls.cache.GetAllBuiltInPositions()); l != 0 {
		t.Errorf("expected builtin positions to be empty, got %d items", l)
	}

	if contents, ok := ls.cache.GetFileContents("file://missing.rego"); ok {
		t.Errorf("expected file contents to be empty, got %s", contents)
	}

	if len(ls.cache.GetAllFiles()) != 0 {
		t.Errorf("expected files to be empty, got %v", ls.cache.GetAllFiles())
	}
}
