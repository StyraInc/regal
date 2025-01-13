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
	"github.com/styrainc/regal/internal/lsp/types"
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

	fileURI := fmt.Sprintf("file://%s/workspace/foo/bar/policy.rego", tmpDir)

	c.SetFileContents(fileURI, "package authz.main.rules")

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

	params, err := ls.fixRenameParams("fix my file!", f, fileURI)
	if err != nil {
		t.Fatalf("failed to fix rename params: %s", err)
	}

	if params.Label != "fix my file!" {
		t.Fatalf("expected label to be 'Fix my file!', got %s", params.Label)
	}

	change, ok := params.Edit.DocumentChanges[0].(types.RenameFile)
	if !ok {
		t.Fatalf("expected document change to be a RenameFile, got %T", params.Edit.DocumentChanges[0])
	}

	if change.Kind != "rename" {
		t.Fatalf("expected kind to be 'rename', got %s", change.Kind)
	}

	if change.OldURI != fileURI {
		t.Fatalf("expected old URI to be %s, got %s", fileURI, change.OldURI)
	}

	if change.NewURI != fmt.Sprintf("file://%s/workspace/authz/main/rules/policy.rego", tmpDir) {
		t.Fatalf("expected new URI to be 'file://%s/workspace/authz/main/rules/policy.rego', got %s", tmpDir, change.NewURI)
	}
}

func TestLanguageServerFixRenameParamsWithConflict(t *testing.T) {
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

	fileURI := fmt.Sprintf("file://%s/workspace/foo/bar/policy.rego", tmpDir)

	c.SetFileContents(fileURI, "package authz.main.rules")

	conflictingFileURI := fmt.Sprintf("file://%s/workspace/authz/main/rules/policy.rego", tmpDir)

	// content of the existing file is irrelevant for this test
	c.SetFileContents(conflictingFileURI, "package authz.main.rules")

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

	params, err := ls.fixRenameParams("fix my file!", f, fileURI)
	if err != nil {
		t.Fatalf("failed to fix rename params: %s", err)
	}

	if params.Label != "fix my file!" {
		t.Fatalf("expected label to be 'Fix my file!', got %s", params.Label)
	}

	if len(params.Edit.DocumentChanges) != 3 {
		t.Fatalf("expected 3 document change, got %d", len(params.Edit.DocumentChanges))
	}

	// check the rename
	change, ok := params.Edit.DocumentChanges[0].(types.RenameFile)
	if !ok {
		t.Fatalf("expected document change to be a RenameFile, got %T", params.Edit.DocumentChanges[0])
	}

	if change.Kind != "rename" {
		t.Fatalf("expected kind to be 'rename', got %s", change.Kind)
	}

	if change.OldURI != fileURI {
		t.Fatalf("expected old URI to be %s, got %s", fileURI, change.OldURI)
	}

	expectedNewURI := fmt.Sprintf("file://%s/workspace/authz/main/rules/policy_1.rego", tmpDir)

	if change.NewURI != expectedNewURI {
		t.Fatalf("expected new URI to be %s, got %s", expectedNewURI, change.NewURI)
	}

	// check the deletes
	deleteChange1, ok := params.Edit.DocumentChanges[1].(types.DeleteFile)
	if !ok {
		t.Fatalf("expected document change to be a DeleteFile, got %T", params.Edit.DocumentChanges[1])
	}

	expectedDeletedURI1 := fmt.Sprintf("file://%s/workspace/foo/bar", tmpDir)
	if deleteChange1.URI != expectedDeletedURI1 {
		t.Fatalf("expected delete URI to be %s, got %s", expectedDeletedURI1, deleteChange1.URI)
	}

	deleteChange2, ok := params.Edit.DocumentChanges[2].(types.DeleteFile)
	if !ok {
		t.Fatalf("expected document change to be a DeleteFile, got %T", params.Edit.DocumentChanges[2])
	}

	expectedDeletedURI2 := fmt.Sprintf("file://%s/workspace/foo", tmpDir)
	if deleteChange2.URI != expectedDeletedURI2 {
		t.Fatalf("expected delete URI to be %s, got %s", expectedDeletedURI2, deleteChange2.URI)
	}
}
