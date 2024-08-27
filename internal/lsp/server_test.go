package lsp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/anderseknert/roast/pkg/encoding"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/styrainc/regal/internal/lsp/types"
)

const mainRegoFileName = "/main.rego"

// defaultTimeout is set based on the investigation done as part of
// https://github.com/StyraInc/regal/issues/931. 20 seconds is 10x the
// maximum time observed for an operation to complete.
const defaultTimeout = 20 * time.Second

const defaultBufferedChannelSize = 5

// InMemoryReadWriteCloser is an in-memory implementation of jsonrpc2.ReadWriteCloser.
type InMemoryReadWriteCloser struct {
	Buffer bytes.Buffer
}

func (m *InMemoryReadWriteCloser) Read(p []byte) (int, error) {
	c, err := m.Buffer.Read(p)

	return c, fmt.Errorf("in memory read failed with error: %w", err)
}

func (m *InMemoryReadWriteCloser) Write(p []byte) (int, error) {
	c, err := m.Buffer.Write(p)

	return c, fmt.Errorf("in memory write failed with error: %w", err)
}

func (*InMemoryReadWriteCloser) Close() error {
	return nil // No-op for in-memory implementation
}

const fileURIScheme = "file://"

// TestLanguageServerSingleFile tests that changes to a single file and Regal config are handled correctly by the
// language server by making updates to both and validating that the correct diagnostics are sent to the client.
//
// This test also ensures that updating the config to point to a non-default engine and capabilities version works
// and causes that engine's builtins to work with completions.
//
//nolint:gocyclo,maintidx
func TestLanguageServerSingleFile(t *testing.T) {
	t.Parallel()

	var err error

	// set up the workspace content with some example rego and regal config
	tempDir := t.TempDir()
	mainRegoURI := fileURIScheme + tempDir + mainRegoFileName

	err = os.MkdirAll(filepath.Join(tempDir, ".regal"), 0o755)
	if err != nil {
		t.Fatalf("failed to create .regal directory: %s", err)
	}

	mainRegoContents := `package main

import rego.v1
allow = true
`

	files := map[string]string{
		"main.rego": mainRegoContents,
		".regal/config.yaml": `
rules:
  idiomatic:
    directory-package-mismatch:
      level: ignore`,
	}

	for f, fc := range files {
		err = os.WriteFile(filepath.Join(tempDir, f), []byte(fc), 0o600)
		if err != nil {
			t.Fatalf("failed to write file %s: %s", f, err)
		}
	}

	// set up the server and client connections
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ls := NewLanguageServer(&LanguageServerOptions{
		ErrorLog: newTestLogger(t),
	})
	go ls.StartDiagnosticsWorker(ctx)
	go ls.StartConfigWorker(ctx)

	receivedMessages := make(chan types.FileDiagnostics, defaultBufferedChannelSize)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		if req.Method == methodTextDocumentPublishDiagnostics {
			var requestData types.FileDiagnostics

			err = encoding.JSON().Unmarshal(*req.Params, &requestData)
			if err != nil {
				t.Fatalf("failed to unmarshal diagnostics: %s", err)
			}

			receivedMessages <- requestData

			return struct{}{}, nil
		}

		t.Fatalf("unexpected request: %v", req)

		return struct{}{}, nil
	}

	connServer, connClient, cleanup := createConnections(ctx, ls.Handle, clientHandler)
	defer cleanup()

	ls.SetConn(connServer)

	// 1. Client sends initialize request
	request := types.InitializeParams{
		RootURI:    fileURIScheme + tempDir,
		ClientInfo: types.Client{Name: "go test"},
	}

	var response types.InitializeResult

	err = connClient.Call(ctx, "initialize", request, &response)
	if err != nil {
		t.Fatalf("failed to send initialize request: %s", err)
	}

	// validate that the server responded with the correct capabilities, and that the correct root URI was set on the
	// server
	if response.Capabilities.DiagnosticProvider.Identifier != "rego" {
		t.Fatalf(
			"expected diagnostic provider identifier to be rego, got %s",
			response.Capabilities.DiagnosticProvider.Identifier,
		)
	}

	if ls.workspaceRootURI != request.RootURI {
		t.Fatalf("expected client root URI to be %s, got %s", request.RootURI, ls.workspaceRootURI)
	}

	// validate that the file contents from the workspace are loaded during the initialize request
	contents, ok := ls.cache.GetFileContents(mainRegoURI)
	if !ok {
		t.Fatalf("expected file contents to be cached")
	}

	if contents != mainRegoContents {
		t.Fatalf("expected file contents to be %s, got %s", mainRegoContents, contents)
	}

	_, ok = ls.cache.GetModule(mainRegoURI)
	if !ok {
		t.Fatalf("expected module to have been parsed and cached for main.rego")
	}

	// 2. Client sends initialized notification
	// no response to the call is expected
	err = connClient.Call(ctx, "initialized", struct{}{}, nil)
	if err != nil {
		t.Fatalf("failed to send initialized notification: %s", err)
	}

	// validate that the client received a diagnostics notification for the file
	timeout := time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			success = testRequestDataCodes(t, requestData, mainRegoURI, []string{"opa-fmt", "use-assignment-operator"})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// 3. Client sends textDocument/didChange notification with new contents for main.rego
	// no response to the call is expected
	err = connClient.Call(ctx, "textDocument/didChange", types.TextDocumentDidChangeParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: mainRegoURI,
		},
		ContentChanges: []types.TextDocumentContentChangeEvent{
			{
				Text: `package main
import rego.v1
allow := true
`,
			},
		},
	}, nil)
	if err != nil {
		t.Fatalf("failed to send didChange notification: %s", err)
	}

	// validate that the client received a new diagnostics notification for the file
	timeout = time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			success = testRequestDataCodes(t, requestData, mainRegoURI, []string{"opa-fmt"})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// 4. Client sends workspace/didChangeWatchedFiles notification with new config
	newConfigContents := `
rules:
  idiomatic:
    directory-package-mismatch:
      level: ignore
  style:
    opa-fmt:
      level: ignore
`

	err = os.WriteFile(filepath.Join(tempDir, ".regal/config.yaml"), []byte(newConfigContents), 0o600)
	if err != nil {
		t.Fatalf("failed to write new config file: %s", err)
	}

	// validate that the client received a new, empty diagnostics notification for the file
	timeout = time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			if requestData.URI != mainRegoURI {
				t.Logf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)

				break
			}

			if len(requestData.Items) != 0 {
				t.Logf("expected 0 diagnostic, got %d", len(requestData.Items))

				break
			}

			success = testRequestDataCodes(t, requestData, mainRegoURI, []string{})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// 5. Client sends new config with an EOPA capabilities file specified.
	newConfigContents = `
rules:
  style:
    opa-fmt:
      level: ignore
capabilities:
  from:
    engine: eopa
    version: v1.23.0
`

	err = os.WriteFile(filepath.Join(tempDir, ".regal/config.yaml"), []byte(newConfigContents), 0o600)
	if err != nil {
		t.Fatalf("failed to write new config file: %s", err)
	}

	// validate that the client received a new, empty diagnostics notification for the file
	timeout = time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			if requestData.URI != mainRegoURI {
				t.Logf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)

				break
			}

			if len(requestData.Items) != 0 {
				t.Logf("expected 0 diagnostic, got %d", len(requestData.Items))

				break
			}

			success = testRequestDataCodes(t, requestData, mainRegoURI, []string{})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// 6. Client sends textDocument/didChange notification with new
	// contents for main.rego no response to the call is expected. We added
	// the start of an EOPA-specific call, so if the capabilities were
	// loaded correctly, we should see a completion later after we ask for
	// it.
	err = connClient.Call(ctx, "textDocument/didChange", types.TextDocumentDidChangeParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: mainRegoURI,
		},
		ContentChanges: []types.TextDocumentContentChangeEvent{
			{
				Text: `package main
import rego.v1
allow := neo4j.q
`,
			},
		},
	}, nil)
	if err != nil {
		t.Fatalf("failed to send didChange notification: %s", err)
	}

	// validate that the client received a new diagnostics notification for the file
	timeout = time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:

			if requestData.URI != mainRegoURI {
				t.Logf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)

				break
			}

			if len(requestData.Items) != 0 {
				t.Logf("expected 0 diagnostic, got %d", len(requestData.Items))

				break
			}

			success = testRequestDataCodes(t, requestData, mainRegoURI, []string{})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// 7. With our new config applied, and the file updated, we can ask the
	// LSP for a completion. We expect to see neo4j.query show up. Since
	// neo4j.query is an EOPA-specific builtin, it should never appear if
	// we're using the normal OPA capabilities file.
	resp := make(map[string]any)
	err = connClient.Call(ctx, "textDocument/completion", types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: mainRegoURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 16,
		},
	}, &resp)
	if err != nil {
		t.Fatalf("failed to send completion notification: %s", err)
	}

	foundNeo4j := false
	for _, itemI := range resp["items"].([]any) {
		item := itemI.(map[string]any)
		t.Logf("completion label: %s", item["label"])
		if item["label"] == "neo4j.query" {
			foundNeo4j = true
		}
	}

	if !foundNeo4j {
		t.Errorf("expected neo4j.query in completion results for neo4j.q")
	}
}

// TestLanguageServerMultipleFiles tests that changes to multiple files are handled correctly. When there are multiple
// files in the workspace, the diagnostics worker also processes aggregate violations, there are also changes to when
// workspace diagnostics are run, this test validates that the correct diagnostics are sent to the client in this
// scenario.
//
// nolint:maintidx,gocyclo
func TestLanguageServerMultipleFiles(t *testing.T) {
	t.Parallel()

	var err error

	// set up the workspace content with some example rego and regal config
	tempDir := t.TempDir()
	authzRegoURI := fileURIScheme + tempDir + "/authz.rego"
	adminsRegoURI := fileURIScheme + tempDir + "/admins.rego"
	ignoredRegoURI := fileURIScheme + tempDir + "/ignored/foo.rego"

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

	err = os.MkdirAll(filepath.Join(tempDir, ".regal"), 0o755)
	if err != nil {
		t.Fatalf("failed to create .regal directory: %s", err)
	}

	err = os.MkdirAll(filepath.Join(tempDir, "ignored"), 0o755)
	if err != nil {
		t.Fatalf("failed to create ignored directory: %s", err)
	}

	for f, fc := range files {
		err = os.WriteFile(filepath.Join(tempDir, f), []byte(fc), 0o600)
		if err != nil {
			t.Fatalf("failed to write file %s: %s", f, err)
		}
	}

	// set up the server and client connections
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ls := NewLanguageServer(&LanguageServerOptions{
		ErrorLog: newTestLogger(t),
	})
	go ls.StartDiagnosticsWorker(ctx)
	go ls.StartConfigWorker(ctx)

	authzFileMessages := make(chan types.FileDiagnostics, defaultBufferedChannelSize)
	adminsFileMessages := make(chan types.FileDiagnostics, defaultBufferedChannelSize)
	ignoredFileMessages := make(chan types.FileDiagnostics, defaultBufferedChannelSize)
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

		switch requestData.URI {
		case authzRegoURI:
			authzFileMessages <- requestData
		case adminsRegoURI:
			adminsFileMessages <- requestData
		case ignoredRegoURI:
			ignoredFileMessages <- requestData
		default:
			t.Logf("unexpected diagnostics for file: %s", requestData.URI)
		}

		return struct{}{}, nil
	}

	connServer, connClient, cleanup := createConnections(ctx, ls.Handle, clientHandler)
	defer cleanup()

	ls.SetConn(connServer)

	// 1. Client sends initialize request
	request := types.InitializeParams{
		RootURI:    fileURIScheme + tempDir,
		ClientInfo: types.Client{Name: "go test"},
	}

	var response types.InitializeResult

	err = connClient.Call(ctx, "initialize", request, &response)
	if err != nil {
		t.Fatalf("failed to send initialize request: %s", err)
	}

	// 2. Client sends initialized notification
	// no response to the call is expected
	err = connClient.Call(ctx, "initialized", struct{}{}, nil)
	if err != nil {
		t.Fatalf("failed to send initialized notification: %s", err)
	}

	// validate that the client received a diagnostics notification for authz.rego
	timeout := time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case diags := <-authzFileMessages:
			success = testRequestDataCodes(t, diags, authzRegoURI, []string{"prefer-package-imports"})
		case <-timeout.C:
			t.Fatalf("timed out waiting for authz.rego diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// validate that the client received a diagnostics notification admins.rego
	timeout = time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case diags := <-adminsFileMessages:
			success = testRequestDataCodes(t, diags, adminsRegoURI, []string{"use-assignment-operator"})
		case <-timeout.C:
			t.Fatalf("timed out waiting for admins.rego diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// 3. Client sends textDocument/didChange notification with new contents for authz.rego
	// no response to the call is expected
	err = connClient.Call(ctx, "textDocument/didChange", types.TextDocumentDidChangeParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: authzRegoURI,
		},
		ContentChanges: []types.TextDocumentContentChangeEvent{
			{
				Text: `package authz

import rego.v1

import data.admins

default allow := false

allow if input.user in admins.users
`,
			},
		},
	}, nil)
	if err != nil {
		t.Fatalf("failed to send didChange notification: %s", err)
	}

	// authz.rego should now have no violations
	timeout = time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case diags := <-authzFileMessages:
			success = testRequestDataCodes(t, diags, authzRegoURI, []string{})
		case <-timeout.C:
			t.Fatalf("timed out waiting for authz.rego diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// we should also receive a diagnostics notification for admins.rego, since it is in the workspace, but it has not
	// been changed, so the violations should be the same.
	timeout = time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case requestData := <-adminsFileMessages:
			success = testRequestDataCodes(t, requestData, adminsRegoURI, []string{"use-assignment-operator"})
		case <-timeout.C:
			t.Fatalf("timed out waiting for admins.rego diagnostics to be sent")
		}

		if success {
			break
		}
	}
}

// https://github.com/StyraInc/regal/issues/679
func TestProcessBuiltinUpdateExitsOnMissingFile(t *testing.T) {
	t.Parallel()

	ls := NewLanguageServer(&LanguageServerOptions{
		ErrorLog: newTestLogger(t),
	})

	err := ls.processHoverContentUpdate(context.Background(), "file://missing.rego", "foo")
	if err != nil {
		t.Fatal(err)
	}

	if l := len(ls.cache.GetAllBuiltInPositions()); l != 0 {
		t.Errorf("expected builtin positions to be empty, got %d items", l)
	}

	contents, ok := ls.cache.GetFileContents("file://missing.rego")
	if ok {
		t.Errorf("expected file contents to be empty, got %s", contents)
	}

	if len(ls.cache.GetAllFiles()) != 0 {
		t.Errorf("expected files to be empty, got %v", ls.cache.GetAllFiles())
	}
}

// TestLanguageServerParentDirConfig tests that regal config is loaded as it is for the
// Regal CLI, and that config files in a parent directory are loaded correctly
// even when the workspace is a child directory.
func TestLanguageServerParentDirConfig(t *testing.T) {
	t.Parallel()

	var err error

	// this is the top level directory for the test
	parentDir := t.TempDir()
	// childDir will be the directory that the client is using as its workspace
	childDirName := "child"
	childDir := filepath.Join(parentDir, childDirName)

	for _, dir := range []string{childDirName, ".regal"} {
		err = os.MkdirAll(filepath.Join(parentDir, dir), 0o755)
		if err != nil {
			t.Fatalf("failed to create %q directory under parent: %s", dir, err)
		}
	}

	mainRegoContents := `package main

import rego.v1
allow := true
`

	files := map[string]string{
		childDirName + mainRegoFileName: mainRegoContents,
		".regal/config.yaml": `rules:
  idiomatic:
    directory-package-mismatch:
      level: ignore
  style:
    opa-fmt:
      level: error
`,
	}

	for f, fc := range files {
		err = os.WriteFile(filepath.Join(parentDir, f), []byte(fc), 0o600)
		if err != nil {
			t.Fatalf("failed to write file %s: %s", f, err)
		}
	}

	// mainRegoFileURI is used throughout the test to refer to the main.rego file
	// and so it is defined here for convenience
	mainRegoFileURI := fileURIScheme + childDir + mainRegoFileName

	// set up the server and client connections
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ls := NewLanguageServer(&LanguageServerOptions{
		ErrorLog: newTestLogger(t),
	})
	go ls.StartDiagnosticsWorker(ctx)
	go ls.StartConfigWorker(ctx)

	receivedMessages := make(chan types.FileDiagnostics, defaultBufferedChannelSize)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		if req.Method == methodTextDocumentPublishDiagnostics {
			var requestData types.FileDiagnostics

			err = encoding.JSON().Unmarshal(*req.Params, &requestData)
			if err != nil {
				t.Fatalf("failed to unmarshal diagnostics: %s", err)
			}

			receivedMessages <- requestData

			return struct{}{}, nil
		}

		t.Logf("unexpected request from server: %v", req)

		return struct{}{}, nil
	}

	connServer, connClient, cleanup := createConnections(ctx, ls.Handle, clientHandler)
	defer cleanup()

	ls.SetConn(connServer)

	// Client sends initialize request
	request := types.InitializeParams{
		RootURI:    fileURIScheme + childDir,
		ClientInfo: types.Client{Name: "go test"},
	}

	var response types.InitializeResult

	err = connClient.Call(ctx, "initialize", request, &response)
	if err != nil {
		t.Fatalf("failed to send initialize request: %s", err)
	}

	if ls.workspaceRootURI != request.RootURI {
		t.Fatalf("expected client root URI to be %s, got %s", request.RootURI, ls.workspaceRootURI)
	}

	// Client sends initialized notification
	// the response to the call is expected to be empty and is ignored
	err = connClient.Call(ctx, "initialized", struct{}{}, nil)
	if err != nil {
		t.Fatalf("failed to send initialized notification: %s", err)
	}

	timeout := time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			success = testRequestDataCodes(t, requestData, mainRegoFileURI, []string{"opa-fmt"})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// User updates config file contents in parent directory that is not
	// part of the workspace
	newConfigContents := `rules:
  idiomatic:
    directory-package-mismatch:
      level: ignore
  style:
    opa-fmt:
      level: ignore
`

	err = os.WriteFile(filepath.Join(parentDir, ".regal/config.yaml"), []byte(newConfigContents), 0o600)
	if err != nil {
		t.Fatalf("failed to write new config file: %s", err)
	}

	// validate that the client received a new, empty diagnostics notification for the file
	timeout = time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			success = testRequestDataCodes(t, requestData, mainRegoFileURI, []string{})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}
}

func testRequestDataCodes(t *testing.T, requestData types.FileDiagnostics, fileURI string, codes []string) bool {
	t.Helper()

	if requestData.URI != fileURI {
		t.Log("expected diagnostics to be sent for", fileURI, "got", requestData.URI)

		return false
	}

	// Extract the codes from requestData.Items
	requestCodes := make([]string, len(requestData.Items))
	for i, item := range requestData.Items {
		requestCodes[i] = item.Code
	}

	// Sort both slices
	sort.Strings(requestCodes)
	sort.Strings(codes)

	if !slices.Equal(requestCodes, codes) {
		t.Logf("expected items: %v, got: %v", codes, requestCodes)

		return false
	}

	return true
}

func createConnections(
	ctx context.Context,
	serverHandler, clientHandler func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error),
) (*jsonrpc2.Conn, *jsonrpc2.Conn, func()) {
	netConnServer, netConnClient := net.Pipe()

	connServer := jsonrpc2.NewConn(
		ctx,
		jsonrpc2.NewBufferedStream(netConnServer, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(serverHandler),
	)

	connClient := jsonrpc2.NewConn(
		ctx,
		jsonrpc2.NewBufferedStream(netConnClient, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(clientHandler),
	)

	cleanup := func() {
		_ = netConnServer.Close()
		_ = netConnClient.Close()
		_ = connServer.Close()
		_ = connClient.Close()
	}

	return connServer, connClient, cleanup
}

// NewTestLogger returns an io.Writer that logs to the given testing.T.
func newTestLogger(t *testing.T) io.Writer {
	t.Helper()

	return &testLogger{t: t}
}

type testLogger struct {
	t *testing.T
}

func (tl *testLogger) Write(p []byte) (n int, err error) {
	tl.t.Log(strings.TrimSpace(string(p)))

	return len(p), nil
}
