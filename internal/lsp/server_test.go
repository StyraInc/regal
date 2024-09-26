package lsp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/styrainc/regal/internal/lsp/types"
)

const mainRegoFileName = "/main.rego"

// defaultTimeout is set based on the investigation done as part of
// https://github.com/StyraInc/regal/issues/931. 20 seconds is 10x the
// maximum time observed for an operation to complete.
const defaultTimeout = 20 * time.Second

const defaultBufferedChannelSize = 5

const fileURIScheme = "file://"

// NewTestLogger returns an io.Writer that logs to the given testing.T.
// This is helpful as it can be used to have the server log to the test logger
// in server tests. It is protected from being written to after the test is
// over.
func newTestLogger(t *testing.T) io.Writer {
	t.Helper()

	tl := &testLogger{t: t, open: true}

	// using cleanup ensure that no goroutines attempt to write to the logger
	// after the test has been cleaned up
	t.Cleanup(func() {
		tl.mu.Lock()
		defer tl.mu.Unlock()
		tl.open = false
	})

	return tl
}

type testLogger struct {
	t    *testing.T
	open bool
	mu   sync.RWMutex
}

func (tl *testLogger) Write(p []byte) (n int, err error) {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	if !tl.open {
		return 0, errors.New("cannot log, test is over")
	}

	tl.t.Log(strings.TrimSpace(string(p)))

	return len(p), nil
}

func createAndInitServer(
	ctx context.Context,
	logger io.Writer,
	tempDir string,
	files map[string]string,
	clientHandler func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error),
) (
	*LanguageServer,
	*jsonrpc2.Conn,
	error,
) {
	var err error

	for f, fc := range files {
		err = os.MkdirAll(filepath.Dir(filepath.Join(tempDir, f)), 0o755)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create directory: %w", err)
		}

		err = os.WriteFile(filepath.Join(tempDir, f), []byte(fc), 0o600)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to write file: %w", err)
		}
	}

	// set up the server and client connections
	ls := NewLanguageServer(ctx, &LanguageServerOptions{
		ErrorLog: logger,
	})

	go ls.StartDiagnosticsWorker(ctx)
	go ls.StartConfigWorker(ctx)

	netConnServer, netConnClient := net.Pipe()

	connServer := jsonrpc2.NewConn(
		ctx,
		jsonrpc2.NewBufferedStream(netConnServer, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(ls.Handle),
	)

	connClient := jsonrpc2.NewConn(
		ctx,
		jsonrpc2.NewBufferedStream(netConnClient, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(clientHandler),
	)

	go func() {
		<-ctx.Done()
		// we need only close the pipe connections as the jsonrpc2.Conn accept
		// the ctx
		_ = netConnClient.Close()
		_ = netConnServer.Close()
	}()

	ls.SetConn(connServer)

	request := types.InitializeParams{
		RootURI:    fileURIScheme + tempDir,
		ClientInfo: types.Client{Name: "go test"},
	}

	var response types.InitializeResult

	err = connClient.Call(ctx, "initialize", request, &response)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize %w", err)
	}

	// 2. Client sends initialized notification
	// no response to the call is expected
	err = connClient.Call(ctx, "initialized", struct{}{}, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to complete initialized %w", err)
	}

	return ls, connClient, nil
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
		t.Logf("waiting for items: %v, got: %v", codes, requestCodes)

		return false
	}

	t.Logf("got expected items")

	return true
}
