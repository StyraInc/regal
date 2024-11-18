package lsp

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/styrainc/regal/internal/lsp/types"
)

// TestLanguageServerMultipleFiles tests that changes to multiple files are handled correctly. When there are multiple
// files in the workspace, the diagnostics worker also processes aggregate violations, there are also changes to when
// workspace diagnostics are run, this test validates that the correct diagnostics are sent to the client in this
// scenario.
//
// nolint:maintidx
func TestLanguageServerMultipleFiles(t *testing.T) {
	t.Parallel()

	// set up the workspace content with some example rego and regal config
	tempDir := t.TempDir()

	files := map[string]string{
		"authz.rego": `package authz

import rego.v1

import data.admins.users

default allow := false

allow if input.user in users
`,
		"admins.rego": `package admins

import rego.v1

users = {"alice", "bob"}
`,
		"ignored/foo.rego": `package ignored

foo = 1
`,
		".regal/config.yaml": `
rules:
  idiomatic:
    directory-package-mismatch:
      level: ignore
ignore:
  files:
    - ignored/*.rego
`,
	}

	logger := newTestLogger(t)

	messages := createMessageChannels(files)

	clientHandler := createClientHandler(t, logger, messages)

	// set up the server and client connections
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ls, connClient, err := createAndInitServer(ctx, logger, tempDir, files, clientHandler)
	if err != nil {
		t.Fatalf("failed to create and init language server: %s", err)
	}

	timeout := time.NewTimer(determineTimeout())
	defer timeout.Stop()

	// wait for the aggregate data to be set, required for correct lint in next
	// step
	for success := false; !success; {
		select {
		default:
			uri := "file://" + filepath.Join(tempDir, "admins.rego")

			aggs := ls.cache.GetFileAggregates(uri)
			if len(aggs) > 0 {
				success = true

				break // don't sleep
			}

			time.Sleep(500 * time.Millisecond)
		case <-timeout.C:
			t.Fatalf("timed out waiting admin aggregates to be set")
		}
	}

	timeout.Reset(determineTimeout())

	// validate that the client received a diagnostics notification for authz.rego
	for success := false; !success; {
		select {
		case violations := <-messages["authz.rego"]:
			if !slices.Contains(violations, "prefer-package-imports") {
				t.Logf("waiting for violations to contain prefer-package-imports, have: %v", violations)

				continue
			}

			success = true
		case <-timeout.C:
			t.Fatalf("timed out waiting for authz.rego diagnostics to be sent")
		}
	}

	// validate that the client received a diagnostics notification for admins.rego
	timeout.Reset(determineTimeout())

	for success := false; !success; {
		select {
		case violations := <-messages["admins.rego"]:
			if !slices.Contains(violations, "use-assignment-operator") {
				t.Logf("waiting for violations to contain use-assignment-operator, have: %v", violations)

				continue
			}

			success = true
		case <-timeout.C:
			t.Fatalf("timed out waiting for admins.rego diagnostics to be sent")
		}
	}

	// 3. Client sends textDocument/didChange notification with new contents
	// for authz.rego no response to the call is expected
	if err := connClient.Notify(ctx, "textDocument/didChange", types.TextDocumentDidChangeParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: fileURIScheme + filepath.Join(tempDir, "authz.rego"),
		},
		ContentChanges: []types.TextDocumentContentChangeEvent{
			{
				Text: `package authz

import rego.v1

import data.admins # fixes prefer-package-imports

default allow := false

# METADATA
# description: Allow only admins
# entrypoint: true # fixes no-defined-entrypoint
allow if input.user in admins.users
`,
			},
		},
	}, nil); err != nil {
		t.Fatalf("failed to send didChange notification: %s", err)
	}

	// authz.rego should now have no violations
	timeout.Reset(determineTimeout())

	for success := false; !success; {
		select {
		case violations := <-messages["authz.rego"]:
			if len(violations) > 0 {
				t.Logf("waiting for violations to be empty for authz.rego, have: %v", violations)

				continue
			}

			success = true
		case <-timeout.C:
			t.Fatalf("timed out waiting for authz.rego diagnostics to be sent")
		}
	}
}
