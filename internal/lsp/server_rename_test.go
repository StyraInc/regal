package lsp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/fixer/fixes"
)

func TestLanguageServerFixRenameParams(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmpDir, "workspace/foo/bar"), 0o755); err != nil {
		t.Fatalf("failed to create directory: %s", err)
	}

	ctx := context.Background()

	logger := newTestLogger(t)

	ls := NewLanguageServer(
		ctx,
		&LanguageServerOptions{LogWriter: logger, LogLevel: log.LevelDebug},
	)

	c := cache.NewCache()
	f := &fixes.DirectoryPackageMismatch{}

	fileURL := fmt.Sprintf("file://%s/workspace/foo/bar/policy.rego", tmpDir)

	c.SetFileContents(fileURL, "package authz.main.rules")

	ls.clientIdentifier = clients.IdentifierVSCode
	ls.workspaceRootURI = fmt.Sprintf("file://%s/workspace", tmpDir)
	ls.cache = c
	ls.loadedConfig = &config.Config{
		Rules: map[string]config.Category{
			"idiomatic": {
				"directory-package-mismatch": config.Rule{
					Level: "ignore",
					Extra: map[string]any{
						"exclude-test-suffix": true,
					},
				},
			},
		},
	}

	params, err := ls.fixRenameParams("fix my file!", f, fileURL)
	if err != nil {
		t.Fatalf("failed to fix rename params: %s", err)
	}

	if params.Label != "fix my file!" {
		t.Fatalf("expected label to be 'Fix my file!', got %s", params.Label)
	}

	if len(params.Edit.DocumentChanges) != 1 {
		t.Fatalf("expected 1 document change, got %d", len(params.Edit.DocumentChanges))
	}

	change := params.Edit.DocumentChanges[0]

	if change.Kind != "rename" {
		t.Fatalf("expected kind to be 'rename', got %s", change.Kind)
	}

	if change.OldURI != fileURL {
		t.Fatalf("expected old URI to be %s, got %s", fileURL, change.OldURI)
	}

	expectedNewURI := fmt.Sprintf("file://%s/workspace/authz/main/rules/policy.rego", tmpDir)

	if change.NewURI != expectedNewURI {
		t.Fatalf("expected new URI to be %s, got %s", expectedNewURI, change.NewURI)
	}
}
