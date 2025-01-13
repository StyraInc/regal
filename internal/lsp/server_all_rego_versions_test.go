package lsp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/internal/lsp/uri"
	"github.com/styrainc/regal/pkg/config"
)

func TestAllRegoVersions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		FileKey         string
		ExpectedVersion ast.RegoVersion
		DiskContents    map[string]string
	}{
		"unknown version": {
			FileKey: "foo/bar.rego",
			DiskContents: map[string]string{
				"foo/bar.rego":       "package foo",
				".regal/config.yaml": "",
			},
			ExpectedVersion: ast.RegoV1,
		},
		"version set in project config": {
			FileKey: "foo/bar.rego",
			DiskContents: map[string]string{
				"foo/bar.rego": "package foo",
				".regal/config.yaml": `
project:
  rego-version: 0
`,
			},
			ExpectedVersion: ast.RegoV0,
		},
		"version set in root config": {
			FileKey: "foo/bar.rego",
			DiskContents: map[string]string{
				"foo/bar.rego": "package foo",
				".regal/config.yaml": `
project:
  rego-version: 1
  roots:
  - path: foo
    rego-version: 0
`,
			},
			ExpectedVersion: ast.RegoV0,
		},
		"version set in manifest": {
			FileKey: "foo/bar.rego",
			DiskContents: map[string]string{
				"foo/bar.rego":       "package foo",
				"foo/.manifest":      `{"rego_version": 0}`,
				".regal/config.yaml": ``,
			},
			ExpectedVersion: ast.RegoV0,
		},
		"version set in manifest, overridden by config": {
			FileKey: "foo/bar.rego",
			DiskContents: map[string]string{
				"foo/bar.rego":  "package foo",
				"foo/.manifest": `{"rego_version": 1}`,
				".regal/config.yaml": `
project:
  roots:
  - path: foo
    rego-version: 0
`,
			},
			ExpectedVersion: ast.RegoV0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			td := t.TempDir()

			// init the state on disk
			for f, c := range tc.DiskContents {
				dir := filepath.Dir(f)

				if err := os.MkdirAll(filepath.Join(td, dir), 0o755); err != nil {
					t.Fatalf("failed to create directory %s: %s", dir, err)
				}

				if err := os.WriteFile(filepath.Join(td, f), []byte(c), 0o600); err != nil {
					t.Fatalf("failed to write file %s: %s", f, err)
				}
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ls := NewLanguageServer(ctx, &LanguageServerOptions{LogWriter: newTestLogger(t), LogLevel: log.LevelDebug})
			ls.workspaceRootURI = uri.FromPath(clients.IdentifierGeneric, td)

			// have the server load the config
			go ls.StartConfigWorker(ctx)

			configFile, err := config.FindConfig(td)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			ls.configWatcher.Watch(configFile.Name())

			// wait for ls.loadedConfig to be set
			timeout := time.NewTimer(determineTimeout())
			defer timeout.Stop()

			for success := false; !success; {
				select {
				default:
					if ls.getLoadedConfig() != nil {
						success = true

						break
					}

					time.Sleep(500 * time.Millisecond)
				case <-timeout.C:
					t.Fatalf("timed out waiting for config to be set")
				}
			}

			// check it has the correct version for the file of interest
			fileURI := uri.FromPath(clients.IdentifierGeneric, filepath.Join(td, tc.FileKey))

			version := ls.determineVersionForFile(fileURI)

			if version != tc.ExpectedVersion {
				t.Errorf("expected version %v, got %v", tc.ExpectedVersion, version)
			}
		})
	}
}
