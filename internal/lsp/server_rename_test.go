package lsp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/testutil"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/fixer/fixes"
)

func TestLanguageServerFixRenameParams(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testutil.MustMkdirAll(t, tmpDir, "workspace", "foo", "bar")

	ctx := context.Background()

	logger := newTestLogger(t)

	ls := NewLanguageServer(
		ctx,
		&LanguageServerOptions{LogWriter: logger, LogLevel: log.LevelDebug},
	)

	c := cache.NewCache()
	f := &fixes.DirectoryPackageMismatch{}

	rootURI := fmt.Sprintf("file://%s/workspace", tmpDir)
	fileURI := rootURI + "/foo/bar/policy.rego"

	c.SetFileContents(fileURI, "package authz.main.rules")

	ls.clientIdentifier = clients.IdentifierVSCode
	ls.workspaceRootURI = rootURI
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

	ctx := context.Background()
	logger := newTestLogger(t)
	tmpDir := t.TempDir()

	testutil.MustMkdirAll(t, tmpDir, "workspace", "foo", "bar")

	ls := NewLanguageServer(
		ctx,
		&LanguageServerOptions{LogWriter: logger, LogLevel: log.LevelDebug},
	)

	c := cache.NewCache()
	f := &fixes.DirectoryPackageMismatch{}

	rootURI := fmt.Sprintf("file://%s/workspace", tmpDir)
	fileURI := rootURI + "/foo/bar/policy.rego"

	c.SetFileContents(fileURI, "package authz.main.rules")

	conflictingFileURI := fmt.Sprintf("file://%s/workspace/authz/main/rules/policy.rego", tmpDir)

	// content of the existing file is irrelevant for this test
	c.SetFileContents(conflictingFileURI, "package authz.main.rules")

	ls.clientIdentifier = clients.IdentifierVSCode
	ls.workspaceRootURI = rootURI
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

func TestLanguageServerFixRenameParamsWhenTargetOutsideRoot(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := newTestLogger(t)
	tmpDir := t.TempDir()

	testutil.MustMkdirAll(t, tmpDir, "workspace", "foo", "bar")

	// this will have regal find a root in the parent dir, which means the file
	// is moved relative to a dir above the workspace.
	err := os.WriteFile(filepath.Join(tmpDir, ".regal.yaml"), []byte{}, 0o600)
	if err != nil {
		t.Fatalf("failed to write config file: %s", err)
	}

	ls := NewLanguageServer(
		ctx,
		&LanguageServerOptions{LogWriter: logger, LogLevel: log.LevelDebug},
	)

	c := cache.NewCache()
	f := &fixes.DirectoryPackageMismatch{}

	// the root where the client stated the workspace is
	rootURI := fmt.Sprintf("file://%s/workspace/", tmpDir)
	fileURI := rootURI + "foo/bar/policy.rego"

	c.SetFileContents(fileURI, "package authz.main.rules")

	ls.clientIdentifier = clients.IdentifierVSCode
	// this is what would be set if a config file were in the parent instead
	ls.workspaceRootURI = rootURI
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

	_, err = ls.fixRenameParams("fix my file!", f, fileURI)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "cannot move file out of workspace root") {
		t.Fatalf("expected error to contain 'cannot move file out of workspace root', got %s", err)
	}
}
