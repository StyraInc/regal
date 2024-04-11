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

	receivedMessages := make(chan jsonrpc2.Request, 1)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		if req.Method == methodTextDocumentPublishDiagnostics {
			receivedMessages <- *req

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
	select {
	case request := <-receivedMessages:
		if request.Method != methodTextDocumentPublishDiagnostics {
			t.Fatalf("expected diagnostics to be sent, got %v", request)
		}

		// validate that the diagnostics are correct
		var requestData FileDiagnostics

		err = json.Unmarshal(*request.Params, &requestData)
		if err != nil {
			t.Fatalf("failed to unmarshal diagnostics: %s", err)
		}

		if requestData.URI != mainRegoURI {
			t.Fatalf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)
		}

		if len(requestData.Items) != 2 {
			t.Fatalf("expected 2 diagnostics, got %d, %v", len(requestData.Items), requestData)
		}

		expectedItems := map[string]bool{
			"opa-fmt":                 false,
			"use-assignment-operator": false,
		}

		for _, item := range requestData.Items {
			expectedItems[item.Code] = true
		}

		for item, found := range expectedItems {
			if !found {
				t.Fatalf("expected diagnostic %s to be found", item)
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for file diagnostics to be sent")
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
	select {
	case request := <-receivedMessages:
		if request.Method != methodTextDocumentPublishDiagnostics {
			t.Fatalf("expected diagnostics to be sent, got %v", request)
		}

		// validate that the diagnostics are correct
		var requestData FileDiagnostics

		err = json.Unmarshal(*request.Params, &requestData)
		if err != nil {
			t.Fatalf("failed to unmarshal diagnostics: %s", err)
		}

		if requestData.URI != mainRegoURI {
			t.Fatalf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)
		}

		if len(requestData.Items) != 1 {
			t.Fatalf("expected 1 diagnostic, got %d", len(requestData.Items))
		}

		if requestData.Items[0].Code != "opa-fmt" {
			t.Fatalf("expected diagnostic to be opa-fmt, got %s", requestData.Items[0].Code)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for file diagnostics to be sent")
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
	select {
	case request := <-receivedMessages:
		if request.Method != methodTextDocumentPublishDiagnostics {
			t.Fatalf("expected diagnostics to be sent, got %v", request)
		}

		// validate that the diagnostics are correct
		var requestData FileDiagnostics

		err = json.Unmarshal(*request.Params, &requestData)
		if err != nil {
			t.Fatalf("failed to unmarshal diagnostics: %s", err)
		}

		if requestData.URI != mainRegoURI {
			t.Fatalf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)
		}

		if len(requestData.Items) != 0 {
			t.Fatalf("expected 0 diagnostic, got %d", len(requestData.Items))
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for file diagnostics to be sent")
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

	authzFileMessages := make(chan FileDiagnostics, 1)
	adminsFileMessages := make(chan FileDiagnostics, 1)
	ignoredFileMessages := make(chan FileDiagnostics, 1)
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

	// validate that the client received a diagnostics notification authz.rego
	select {
	case requestData := <-authzFileMessages:
		if requestData.URI != authzRegoURI {
			t.Fatalf("expected diagnostics to be sent for authz.rego, got %s", requestData.URI)
		}

		if len(requestData.Items) != 1 {
			t.Fatalf("expected 1 diagnostics, got %d", len(requestData.Items))
		}

		if requestData.Items[0].Code != "prefer-package-imports" {
			t.Fatalf("expected diagnostic to be prefer-package-imports, got %s", requestData.Items[0].Code)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for authz.rego diagnostics to be sent")
	}

	// validate that the client received a diagnostics notification admins.rego
	select {
	case requestData := <-adminsFileMessages:
		if requestData.URI != adminsRegoURI {
			t.Fatalf("expected diagnostics to be sent for admins.rego, got %s", requestData.URI)
		}

		if len(requestData.Items) != 1 {
			t.Fatalf("expected 1 diagnostics, got %d, %v", len(requestData.Items), requestData)
		}

		if requestData.Items[0].Code != "use-assignment-operator" {
			t.Fatalf("expected diagnostic to be use-assignment-operator, got %s", requestData.Items[0].Code)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for admins.rego diagnostics to be sent")
	}

	select {
	case requestData := <-ignoredFileMessages:
		if requestData.URI != ignoredRegoURI {
			t.Fatalf("expected diagnostics to be sent for ignored/foo.rego, got %s", requestData.URI)
		}

		// this file is ignored, so there should be no diagnostics
		if len(requestData.Items) != 0 {
			t.Fatalf("expected 0 diagnostics, got %d, %v", len(requestData.Items), requestData)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for ignored/foo.rego diagnostics to be sent")
	}

	// 3. Client sends textDocument/didChange notification with new contents for main.rego
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

	// validate that diagnostics are sent for both files since the change was made to a file with an aggregate
	// violation

	// here we wait to receive a diagnostics notification for authz.rego with no diagnostics items, the file diagnostics
	// can arrive first which can still contain the old diagnostics items
	ok := make(chan bool, 1)

	go func() {
		for {
			requestData := <-authzFileMessages
			if requestData.URI != authzRegoURI {
				t.Logf("expected diagnostics to be sent for authz.rego, got %s", requestData.URI)
			}

			if len(requestData.Items) != 0 {
				continue
			}

			ok <- true

			return
		}
	}()

	select {
	case <-ok:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for authz.rego diagnostics to be sent")
	}

	// we should also receive a diagnostics notification for admins.rego, since it is in the workspace, but it has not
	// been changed, so the violations should be the same
	select {
	case requestData := <-adminsFileMessages:
		if requestData.URI != adminsRegoURI {
			t.Fatalf("expected diagnostics to be sent for admins.rego, got %s", requestData.URI)
		}

		// this file is unchanged
		if len(requestData.Items) != 1 {
			t.Fatalf("expected 1 diagnostics, got %d", len(requestData.Items))
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for admins.rego diagnostics to be sent")
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

	receivedMessages := make(chan jsonrpc2.Request, 1)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		if req.Method == methodTextDocumentPublishDiagnostics {
			receivedMessages <- *req

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

	// validate that the client received a diagnostics notification for the file
	select {
	case request := <-receivedMessages:
		// validate that the diagnostics are correct
		var requestData FileDiagnostics

		err = json.Unmarshal(*request.Params, &requestData)
		if err != nil {
			t.Fatalf("failed to unmarshal diagnostics: %s", err)
		}

		if requestData.URI != mainRegoFileURI {
			t.Fatalf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)
		}

		if len(requestData.Items) != 1 {
			t.Fatalf("expected 1 diagnostics, got %d, %v", len(requestData.Items), requestData)
		}

		if requestData.Items[0].Code != "opa-fmt" {
			t.Fatalf("expected diagnostic to be opa-fmt, got %s", requestData.Items[0].Code)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for file diagnostics to be sent")
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
	select {
	case request := <-receivedMessages:
		// validate that the diagnostics are correct
		var requestData FileDiagnostics

		err = json.Unmarshal(*request.Params, &requestData)
		if err != nil {
			t.Fatalf("failed to unmarshal diagnostics: %s", err)
		}

		if requestData.URI != mainRegoFileURI {
			t.Fatalf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)
		}

		if len(requestData.Items) != 0 {
			t.Fatalf("expected 0 diagnostic, got %d", len(requestData.Items))
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for file diagnostics to be sent")
	}
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
