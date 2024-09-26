package lsp

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/anderseknert/roast/pkg/encoding"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/util"
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

	messages := make(map[string]chan []string)
	for _, file := range util.Keys(files) {
		messages[file] = make(chan []string, 10)
	}

	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		if req.Method != "textDocument/publishDiagnostics" {
			t.Log("unexpected request method:", req.Method)

			return struct{}{}, nil
		}

		var requestData types.FileDiagnostics

		err = encoding.JSON().Unmarshal(*req.Params, &requestData)
		if err != nil {
			t.Fatalf("failed to unmarshal diagnostics: %s", err)
		}

		violations := make([]string, len(requestData.Items))
		for i, item := range requestData.Items {
			violations[i] = item.Code
		}

		slices.Sort(violations)

		messages[filepath.Base(requestData.URI)] <- violations

		return struct{}{}, nil
	}

	// set up the server and client connections
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, connClient, err := createAndInitServer(ctx, newTestLogger(t), tempDir, files, clientHandler)
	if err != nil {
		t.Fatalf("failed to create and init language server: %s", err)
	}

	// validate that the client received a diagnostics notification for authz.rego
	timeout := time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case violations := <-messages["authz.rego"]:
			if !slices.Contains(violations, "prefer-package-imports") {
				t.Logf("waiting for violations to contain prefer-package-imports")

				continue
			}

			success = true
		case <-timeout.C:
			t.Fatalf("timed out waiting for authz.rego diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// validate that the client received a diagnostics notification admins.rego
	timeout.Reset(defaultTimeout)

	for {
		var success bool
		select {
		case violations := <-messages["admins.rego"]:
			if !slices.Contains(violations, "use-assignment-operator") {
				t.Logf("waiting for violations to contain use-assignment-operator")

				continue
			}

			success = true
		case <-timeout.C:
			t.Fatalf("timed out waiting for admins.rego diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// 3. Client sends textDocument/didChange notification with new contents for authz.rego
	// no response to the call is expected
	if err := connClient.Call(ctx, "textDocument/didChange", types.TextDocumentDidChangeParams{
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
	timeout.Reset(defaultTimeout)

	for {
		var success bool
		select {
		case violations := <-messages["authz.rego"]:
			if len(violations) > 0 {
				t.Logf("waiting for violations to be empty for authz.rego")

				continue
			}

			success = true
		case <-timeout.C:
			t.Fatalf("timed out waiting for authz.rego diagnostics to be sent")
		}

		if success {
			break
		}
	}
}
