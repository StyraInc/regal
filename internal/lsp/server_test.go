package lsp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anderseknert/roast/pkg/encoding"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/util"
)

const mainRegoFileName = "/main.rego"

// defaultTimeout is set based on the investigation done as part of
// https://github.com/StyraInc/regal/issues/931. 20 seconds is 10x the
// maximum time observed for an operation to complete.
const defaultTimeout = 20 * time.Second

const defaultBufferedChannelSize = 5

// determineTimeout returns a timeout duration based on whether
// the test suite is running with race detection, if so, a more permissive
// timeout is used.
func determineTimeout() time.Duration {
	if isRaceEnabled() {
		// based on the upper bound here, 20x slower
		// https://go.dev/doc/articles/race_detector#Runtime_Overheads
		return defaultTimeout * 20
	}

	return defaultTimeout
}

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
	t *testing.T,
	ctx context.Context,
	logger io.Writer,
	tempDir string,
	clientHandler func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error),
) (
	*LanguageServer,
	*jsonrpc2.Conn,
) {
	t.Helper()

	// This is set due to eventing being so slow in go test -race that we
	// get flakes. TODO, work out how to avoid needing this in lsp tests.
	pollingInterval := time.Duration(0)
	if isRaceEnabled() {
		pollingInterval = 10 * time.Second
	}

	// set up the server and client connections
	ls := NewLanguageServer(ctx, &LanguageServerOptions{
		LogWriter:                logger,
		LogLevel:                 log.LevelDebug,
		WorkspaceDiagnosticsPoll: pollingInterval,
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

	if err := connClient.Call(ctx, "initialize", request, &response); err != nil {
		t.Fatalf("failed to initialize: %s", err)
	}

	// 2. Client sends initialized notification
	// no response to the call is expected
	if err := connClient.Call(ctx, "initialized", struct{}{}, nil); err != nil {
		t.Fatalf("failed to complete initialized: %v", err)
	}

	return ls, connClient
}

func createClientHandler(
	t *testing.T,
	logger io.Writer,
	messages map[string]chan []string,
) func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	t.Helper()

	return func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		if req.Method != "textDocument/publishDiagnostics" {
			fmt.Fprintln(logger, "createClientHandler: unexpected request method:", req.Method)

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

		fileBase := filepath.Base(requestData.URI)
		fmt.Fprintln(logger, "createClientHandler: queue", fileBase, len(messages[fileBase]))

		select {
		case messages[fileBase] <- violations:
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout writing to messages channel for %s", fileBase)
		}

		return struct{}{}, nil
	}
}

func createMessageChannels(files map[string]string) map[string]chan []string {
	messages := make(map[string]chan []string)
	for _, file := range util.Keys(files) {
		messages[file] = make(chan []string, 10)
	}

	return messages
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
