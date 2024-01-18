package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sourcegraph/jsonrpc2"
)

// InMemoryReadWriteCloser is an in-memory implementation of jsonrpc2.ReadWriteCloser.
type InMemoryReadWriteCloser struct {
	Buffer bytes.Buffer
}

func (m *InMemoryReadWriteCloser) Read(p []byte) (int, error) {
	return m.Buffer.Read(p)
}

func (m *InMemoryReadWriteCloser) Write(p []byte) (int, error) {
	return m.Buffer.Write(p)
}

func (m *InMemoryReadWriteCloser) Close() error {
	return nil // No-op for in-memory implementation
}

// TestLanguageServerSingleFile tests that changes to a single file and Regal config are handled correctly by the
// language server my making updates to both and validating that the correct diagnostics are sent to the client.
func TestLanguageServerSingleFileWithConfig(t *testing.T) {
	var err error

	// set up the workspace content with some example rego and regal config
	tempDir := t.TempDir()

	err = os.MkdirAll(filepath.Join(tempDir, ".regal"), 0755)
	if err != nil {
		t.Fatalf("failed to create .regal directory: %s", err)
	}

	mainRegoContents := `package main
allow = true
`

	files := map[string]string{
		"main.rego":          mainRegoContents,
		".regal/config.yaml": ``,
	}

	for f, fc := range files {
		err = os.WriteFile(filepath.Join(tempDir, f), []byte(fc), 0644)
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
	ls.StartDiagnosticsWorker(ctx)

	receivedMessages := make(chan jsonrpc2.Request, 1)
	testHandler := func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
		if req.Method == "textDocument/publishDiagnostics" {
			receivedMessages <- *req
			return nil, nil
		}
		t.Fatalf("unexpected request: %v", req)
		return nil, nil
	}

	netConnServer, netConnClient := net.Pipe()
	defer netConnServer.Close()
	defer netConnClient.Close()
	connServer := jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(netConnServer, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.HandlerWithError(ls.Handle))
	connClient := jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(netConnClient, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.HandlerWithError(testHandler))
	defer connServer.Close()
	defer connClient.Close()
	ls.SetConn(connServer)

	// 1. Client sends initialize request
	request := InitializeParams{
		RootURI: "file://" + tempDir,
	}

	var response InitializeResult
	err = connClient.Call(ctx, "initialize", request, &response)
	if err != nil {
		t.Fatalf("failed to send initialize request: %s", err)
	}

	// validate that the server responded with the correct capabilities, and that the correct root URI was set on the
	// server
	if response.Capabilities.DiagnosticProvider.Identifier != "rego" {
		t.Fatalf("expected diagnostic provider identifier to be rego, got %s", response.Capabilities.DiagnosticProvider.Identifier)
	}
	if ls.clientRootURI != request.RootURI {
		t.Fatalf("expected client root URI to be %s, got %s", request.RootURI, ls.clientRootURI)
	}

	// validate that the file contents from the workspace are loaded during the initialize request
	contents, ok := ls.cache.GetFileContents("file://" + tempDir + "/main.rego")
	if !ok {
		t.Fatalf("expected file contents to be cached")
	}
	if contents != mainRegoContents {
		t.Fatalf("expected file contents to be %s, got %s", mainRegoContents, contents)
	}
	_, ok = ls.cache.GetModule("file://" + tempDir + "/main.rego")
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
		if request.Method != "textDocument/publishDiagnostics" {
			t.Fatalf("expected diagnostics to be sent, got %v", request)
		}

		// validate that the diagnostics are correct
		var requestData FileDiagnostics
		err = json.Unmarshal(*request.Params, &requestData)
		if err != nil {
			t.Fatalf("failed to unmarshal diagnostics: %s", err)
		}

		if requestData.URI != "file://"+tempDir+"/main.rego" {
			t.Fatalf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)
		}
		if len(requestData.Items) != 2 {
			t.Fatalf("expected 2 diagnostics, got %d", len(requestData.Items))
		}
		expectedItems := map[string]bool{
			"opa-fmt":                 false,
			"use-assignment-operator": false,
		}
		for _, item := range requestData.Items {
			t.Log(item.Code)
			expectedItems[item.Code.Value] = true
		}
		for item, found := range expectedItems {
			if !found {
				t.Fatalf("expected diagnostic %s to be found", item)
			}
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for file diagnostics to be sent")
	}

	// 3. Client sends textDocument/didChange notification with new contents for main.rego
	// no response to the call is expected
	err = connClient.Call(ctx, "textDocument/didChange", TextDocumentDidChangeParams{
		TextDocument: TextDocumentIdentifier{
			URI: "file://" + tempDir + "/main.rego",
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{
				Text: `package main
allow := true
`,
			},
		},
	}, nil)

	// validate that the client received a new diagnostics notification for the file
	select {
	case request := <-receivedMessages:
		if request.Method != "textDocument/publishDiagnostics" {
			t.Fatalf("expected diagnostics to be sent, got %v", request)
		}

		// validate that the diagnostics are correct
		var requestData FileDiagnostics
		err = json.Unmarshal(*request.Params, &requestData)
		if err != nil {
			t.Fatalf("failed to unmarshal diagnostics: %s", err)
		}

		if requestData.URI != "file://"+tempDir+"/main.rego" {
			t.Fatalf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)
		}
		if len(requestData.Items) != 1 {
			t.Fatalf("expected 1 diagnostic, got %d", len(requestData.Items))
		}
		if requestData.Items[0].Code.Value != "opa-fmt" {
			t.Fatalf("expected diagnostic to be opa-fmt, got %s", requestData.Items[0].Code.Value)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for file diagnostics to be sent")
	}

	// 4. Client sends workspace/didChangeWatchedFiles notification with new config
	newConfigContents := `
rules:
  style:
    opa-fmt:
      level: ignore
`
	err = os.WriteFile(filepath.Join(tempDir, ".regal/config.yaml"), []byte(newConfigContents), 0644)
	if err != nil {
		t.Fatalf("failed to write new config file: %s", err)
	}
	err = connClient.Call(ctx, "workspace/didChangeWatchedFiles", WorkspaceDidChangeWatchedFilesParams{
		Changes: []FileEvent{
			{
				Type: 1,
				URI:  "file://" + tempDir + "/.regal/config.yaml",
			},
		},
	}, nil)

	// validate that the client received a new, empty diagnostics notification for the file
	select {
	case request := <-receivedMessages:
		if request.Method != "textDocument/publishDiagnostics" {
			t.Fatalf("expected diagnostics to be sent, got %v", request)
		}

		// validate that the diagnostics are correct
		var requestData FileDiagnostics
		err = json.Unmarshal(*request.Params, &requestData)
		if err != nil {
			t.Fatalf("failed to unmarshal diagnostics: %s", err)
		}

		if requestData.URI != "file://"+tempDir+"/main.rego" {
			t.Fatalf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)
		}
		if len(requestData.Items) != 0 {
			t.Fatalf("expected 1 diagnostic, got %d", len(requestData.Items))
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for file diagnostics to be sent")
	}

}

// TestLanguageServerMultipleFiles tests that changes to multiple files are handled correctly. When there are multiple
// files in the workspace, the diagnostics worker also processes aggregate violations, there are also changes to when
// workspace diagnostics are run, this test validates that the correct diagnostics are sent to the client in this
// scenario.
func TestLanguageServerMultipleFiles(t *testing.T) {
	var err error

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

users = {"alice", "bob"}
`,
	}

	for f, fc := range files {
		err = os.WriteFile(filepath.Join(tempDir, f), []byte(fc), 0644)
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
	ls.StartDiagnosticsWorker(ctx)

	authzFileMessages := make(chan FileDiagnostics, 1)
	adminsFileMessages := make(chan FileDiagnostics, 1)
	testHandler := func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
		if req.Method == "textDocument/publishDiagnostics" {
			var requestData FileDiagnostics
			err = json.Unmarshal(*req.Params, &requestData)
			if err != nil {
				t.Fatalf("failed to unmarshal diagnostics: %s", err)
			}

			if requestData.URI == "file://"+tempDir+"/authz.rego" {
				authzFileMessages <- requestData
			}
			if requestData.URI == "file://"+tempDir+"/admins.rego" {
				adminsFileMessages <- requestData
			}

			return nil, nil
		}
		t.Fatalf("unexpected request: %v", req)
		return nil, nil
	}

	netConnServer, netConnClient := net.Pipe()
	defer netConnServer.Close()
	defer netConnClient.Close()
	connServer := jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(netConnServer, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.HandlerWithError(ls.Handle))
	connClient := jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(netConnClient, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.HandlerWithError(testHandler))
	defer connServer.Close()
	defer connClient.Close()
	ls.SetConn(connServer)

	// 1. Client sends initialize request
	request := InitializeParams{
		RootURI: "file://" + tempDir,
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
		if requestData.URI != "file://"+tempDir+"/authz.rego" {
			t.Fatalf("expected diagnostics to be sent for authz.rego, got %s", requestData.URI)
		}
		if len(requestData.Items) != 1 {
			t.Fatalf("expected 1 diagnostics, got %d", len(requestData.Items))
		}
		if requestData.Items[0].Code.Value != "prefer-package-imports" {
			t.Fatalf("expected diagnostic to be prefer-package-imports, got %s", requestData.Items[0].Code.Value)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for authz.rego diagnostics to be sent")
	}

	// validate that the client received a diagnostics notification admins.rego
	select {
	case requestData := <-adminsFileMessages:
		if requestData.URI != "file://"+tempDir+"/admins.rego" {
			t.Fatalf("expected diagnostics to be sent for admins.rego, got %s", requestData.URI)
		}

		if len(requestData.Items) != 1 {
			t.Fatalf("expected 1 diagnostics, got %d", len(requestData.Items))
		}

		if requestData.Items[0].Code.Value != "use-assignment-operator" {
			t.Fatalf("expected diagnostic to be use-assignment-operator, got %s", requestData.Items[0].Code.Value)
		}
	}

	// 3. Client sends textDocument/didChange notification with new contents for main.rego
	// no response to the call is expected
	err = connClient.Call(ctx, "textDocument/didChange", TextDocumentDidChangeParams{
		TextDocument: TextDocumentIdentifier{
			URI: "file://" + tempDir + "/authz.rego",
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

	// validate that diagnostics are sent for both files since the change was made to a file with an aggregate
	// violation

	// here we wait to receive a diagnostics notification for authz.rego with no diagnostics items, the file diagnostics
	// can arrive first which can still contain the old diagnostics items
	ok := make(chan bool, 1)
	go func() {
		for {
			requestData := <-authzFileMessages
			if requestData.URI != "file://"+tempDir+"/authz.rego" {
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
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for authz.rego diagnostics to be sent")
	}

	// we should also receive a diagnostics notification for admins.rego, since it is in the workspace, but it has not
	// been changed, so the violations should be the same
	select {
	case requestData := <-adminsFileMessages:
		if requestData.URI != "file://"+tempDir+"/admins.rego" {
			t.Fatalf("expected diagnostics to be sent for admins.rego, got %s", requestData.URI)
		}

		// this file is unchanged
		if len(requestData.Items) != 1 {
			t.Fatalf("expected 1 diagnostics, got %d", len(requestData.Items))
		}
	}
}