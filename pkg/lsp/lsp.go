package lsp

import (
	"context"
	"os"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	jsonrpc2_ws "github.com/sourcegraph/jsonrpc2/websocket"

	"github.com/styrainc/regal/internal/lsp"
	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/pkg/config"
)

type Handle struct {
	conn *jsonrpc2.Conn
	ls   *lsp.LanguageServer
}

func New(ctx context.Context, ws *websocket.Conn, c *config.Config) (*Handle, error) {
	opts := lsp.LanguageServerOptions{
		LogWriter: os.Stderr,
		LogLevel:  log.LevelOff,
	}
	if os.Getenv("REGAL_DEBUG") != "" {
		opts.LogLevel = log.LevelDebug
	}

	ls := lsp.NewLanguageServerMinimal(ctx, &opts, c)
	jconn := jsonrpc2.NewConn(
		ctx,
		jsonrpc2_ws.NewObjectStream(ws),
		jsonrpc2.HandlerWithError(ls.Handle),
	)
	ls.SetConn(jconn)

	go ls.StartDiagnosticsWorker(ctx)
	go ls.StartHoverWorker(ctx)
	go ls.StartCommandWorker(ctx)

	return &Handle{
		conn: jconn,
		ls:   ls,
	}, nil
}

func (h *Handle) Wait(ctx context.Context) error {
	select {
	case <-h.conn.DisconnectNotify():
		return nil
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck
	}
}

func (h *Handle) Close() error {
	return h.conn.Close() //nolint:wrapcheck
}
