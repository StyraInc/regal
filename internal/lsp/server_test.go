package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sourcegraph/jsonrpc2"
)

const mainRegoFileName = "/main.rego"

const defaultTimeout = 3 * time.Second

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
// language server my making updates to both and validating that the correct diagnostics are sent to the client.
//
//nolint:gocognit,gocyclo,maintidx
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
		"main.rego":          mainRegoContents,
		".regal/config.yaml": ``,
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
		ErrorLog: os.Stderr,
	})
	go ls.StartDiagnosticsWorker(ctx)
	go ls.StartConfigWorker(ctx)

	receivedMessages := make(chan FileDiagnostics, defaultBufferedChannelSize)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		if req.Method == methodTextDocumentPublishDiagnostics {
			var requestData FileDiagnostics

			err = json.Unmarshal(*req.Params, &requestData)
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
	request := InitializeParams{
		RootURI:    fileURIScheme + tempDir,
		ClientInfo: Client{Name: "go test"},
	}

	var response InitializeResult

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

	if ls.clientRootURI != request.RootURI {
		t.Fatalf("expected client root URI to be %s, got %s", request.RootURI, ls.clientRootURI)
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
	err = connClient.Call(ctx, "textDocument/didChange", TextDocumentDidChangeParams{
		TextDocument: TextDocumentIdentifier{
			URI: mainRegoURI,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
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
}

// TestLanguageServerMultipleFiles tests that changes to multiple files are handled correctly. When there are multiple
// files in the workspace, the diagnostics worker also processes aggregate violations, there are also changes to when
// workspace diagnostics are run, this test validates that the correct diagnostics are sent to the client in this
// scenario.
//
//nolint:gocognit,maintidx,gocyclo
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
		ErrorLog: os.Stderr,
	})
	go ls.StartDiagnosticsWorker(ctx)
	go ls.StartConfigWorker(ctx)

	authzFileMessages := make(chan FileDiagnostics, defaultBufferedChannelSize)
	adminsFileMessages := make(chan FileDiagnostics, defaultBufferedChannelSize)
	ignoredFileMessages := make(chan FileDiagnostics, defaultBufferedChannelSize)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		if req.Method == "textDocument/publishDiagnostics" {
			var requestData FileDiagnostics

			err = json.Unmarshal(*req.Params, &requestData)
			if err != nil {
				t.Fatalf("failed to unmarshal diagnostics: %s", err)
			}

			if requestData.URI == authzRegoURI {
				authzFileMessages <- requestData
			}

			if requestData.URI == adminsRegoURI {
				adminsFileMessages <- requestData
			}

			if requestData.URI == ignoredRegoURI {
				ignoredFileMessages <- requestData
			}

			return struct{}{}, nil
		}

		t.Fatalf("unexpected request: %v", req)

		return struct{}{}, nil
	}

	connServer, connClient, cleanup := createConnections(ctx, ls.Handle, clientHandler)
	defer cleanup()

	ls.SetConn(connServer)

	// 1. Client sends initialize request
	request := InitializeParams{
		RootURI:    fileURIScheme + tempDir,
		ClientInfo: Client{Name: "go test"},
	}

	var response InitializeResult

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

	// validate that the client received empty diagnostics for the ignored file
	timeout = time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case diags := <-ignoredFileMessages:
			success = testRequestDataCodes(t, diags, ignoredRegoURI, []string{})
		case <-timeout.C:
			t.Fatalf("timed out waiting for empty ignored file diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// 3. Client sends textDocument/didChange notification with new contents for authz.rego
	// no response to the call is expected
	err = connClient.Call(ctx, "textDocument/didChange", TextDocumentDidChangeParams{
		TextDocument: TextDocumentIdentifier{
			URI: authzRegoURI,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
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
			t.Fatalf("timed out waiting for file diagnostics to be sent")
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
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
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
		ErrorLog: os.Stderr,
	})
	go ls.StartDiagnosticsWorker(ctx)
	go ls.StartConfigWorker(ctx)

	receivedMessages := make(chan FileDiagnostics, defaultBufferedChannelSize)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		if req.Method == methodTextDocumentPublishDiagnostics {
			var requestData FileDiagnostics

			err = json.Unmarshal(*req.Params, &requestData)
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
	request := InitializeParams{
		RootURI:    fileURIScheme + childDir,
		ClientInfo: Client{Name: "go test"},
	}

	var response InitializeResult

	err = connClient.Call(ctx, "initialize", request, &response)
	if err != nil {
		t.Fatalf("failed to send initialize request: %s", err)
	}

	if ls.clientRootURI != request.RootURI {
		t.Fatalf("expected client root URI to be %s, got %s", request.RootURI, ls.clientRootURI)
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

func testRequestDataCodes(t *testing.T, requestData FileDiagnostics, uri string, codes []string) bool {
	t.Helper()

	if requestData.URI != uri {
		t.Log("expected diagnostics to be sent for", uri, "got", requestData.URI)

		return false
	}

	if len(requestData.Items) != len(codes) {
		t.Log("expected", len(codes), "diagnostics, got", len(requestData.Items))

		return false
	}

	for _, v := range codes {
		found := false
		foundItems := make([]string, 0, len(requestData.Items))

		for _, i := range requestData.Items {
			foundItems = append(foundItems, i.Code)

			if i.Code == v {
				found = true

				break
			}
		}

		if !found {
			t.Log("expected diagnostic", v, "not found in", foundItems)

			return false
		}
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
		netConnServer.Close()
		netConnClient.Close()
		connServer.Close()
		connClient.Close()
	}

	return connServer, connClient, cleanup
}
