package lsp

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	jsonrpc2_ws "github.com/sourcegraph/jsonrpc2/websocket"

	"github.com/styrainc/regal/internal/lsp"
)

type Handle struct {
	conn *jsonrpc2.Conn
	ls   *lsp.LanguageServer
}

func New(ctx context.Context, ws *websocket.Conn) (*Handle, error) {
	opts := lsp.LanguageServerOptions{}
	ls := lsp.NewLanguageServer(ctx, &opts)
	jconn := jsonrpc2.NewConn(
		ctx,
		jsonrpc2_ws.NewObjectStream(ws),
		jsonrpc2.HandlerWithError(ls.Handle),
	)
	ls.SetConn(jconn)

	go ls.StartDiagnosticsWorker(ctx)
	go ls.StartHoverWorker(ctx)
	go ls.StartCommandWorker(ctx)
	go ls.StartConfigWorker(ctx)
	go ls.StartWorkspaceStateWorker(ctx)
	go ls.StartTemplateWorker(ctx)
	go ls.StartWebServer(ctx)

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
		return ctx.Err()
	}
}

func (h *Handle) Close() error {
	return h.conn.Close() //nolint:wrapcheck
}
