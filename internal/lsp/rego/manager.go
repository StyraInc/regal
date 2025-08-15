package rego

import (
	"context"
	"errors"

	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/styrainc/regal/internal/lsp/handler"
	"github.com/styrainc/regal/internal/lsp/types"
)

const entrypoint = "data.regal.lsp.main.eval"

var (
	emptyResponse = map[string]any{
		"textDocument/codeAction":        make([]types.CodeAction, 0),
		"textDocument/documentLink":      make([]types.DocumentLink, 0),
		"textDocument/documentHighlight": make([]types.DocumentHighlight, 0),
		"textDocument/documentSymbol":    make([]types.DocumentSymbol, 0),
		"textDocument/completion":        make([]types.CompletionItem, 0),
		"textDocument/codeLens":          make([]types.CodeLens, 0),
		"textDocument/signatureHelp":     nil,
	}

	errIgnored = errors.New("ignored URI")
)

type (
	Providers struct {
		ContextProvider func(uri string, reqs *Requirements) RegalContext
		ContentProvider func(uri string) (string, bool)
		IgnoredProvider func(uri string) bool
	}

	RegoManager struct {
		routes    map[string]Route
		providers Providers
	}

	Route struct {
		handler  regoContextHandler
		requires *Requirements
	}

	regoHandler        = func(context.Context, Providers, *jsonrpc2.Request) (any, error)
	regoContextHandler = func(context.Context, RegalContext, *jsonrpc2.Request) (any, error)
)

func NewRegoManager(ctx context.Context, store storage.Store, prvs Providers) *RegoManager {
	if err := StoreCachedQuery(ctx, entrypoint, store); err != nil {
		panic(err)
	}

	routes := map[string]Route{
		"textDocument/codeAction": {
			handler: textDocument[types.CodeActionParams, []types.CodeAction],
		},
		"textDocument/documentLink": {
			handler: textDocument[types.DocumentLinkParams, []types.DocumentLink],
		},
		"textDocument/documentHighlight": {
			handler: textDocument[types.DocumentHighlightParams, []types.DocumentHighlight],
			requires: &Requirements{
				File: FileRequirements{Lines: true},
			},
		},
		"textDocument/signatureHelp": {
			handler: textDocument[types.SignatureHelpParams, *types.SignatureHelp],
			requires: &Requirements{
				File: FileRequirements{Lines: true},
			},
		},
	}

	return &RegoManager{routes: routes, providers: prvs}
}

func (m *RegoManager) Handle(ctx context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
	if route, ok := m.routes[req.Method]; ok {
		handler := uriCheckHandler(route)

		return handler(ctx, m.providers, req)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: "method not supported: " + req.Method}
}

// uriCheckHandler wraps a regoHandler to check if the URI is ignored before calling the handler.
func uriCheckHandler(route Route) regoHandler {
	return func(ctx context.Context, prvs Providers, req *jsonrpc2.Request) (any, error) {
		uri, err := decodeAndCheckURI(req, prvs.IgnoredProvider)
		if err != nil {
			if errors.Is(err, errIgnored) {
				return emptyResponse[req.Method], nil
			}
			return nil, err
		}

		return route.handler(ctx, prvs.ContextProvider(uri, route.requires), req)
	}
}

// textDocument is a handler that requires TextDocumentParams (i.e. a document URI)
// embedded in parameter of type P, returning a result of type R.
func textDocument[P, R any](ctx context.Context, rctx RegalContext, req *jsonrpc2.Request) (any, error) {
	params, err := decodeParams[P](req)
	if err != nil {
		return nil, err
	}

	result, err := QueryEval[P, R](ctx, entrypoint, NewInputWithMethod(req.Method, rctx, params))
	if err != nil {
		return nil, err
	}

	// For now we just unwrap the LSP response here, but may use other fields in the future.
	// In particular, we'll likely want to allow Rego handlers to return detailed error messages.
	return result.Response, nil
}

func decodeAndCheckURI(req *jsonrpc2.Request, ignored func(string) bool) (string, error) {
	tdp, err := decodeParams[types.TextDocumentParams](req)
	if err != nil {
		return "", err
	}

	if ignored(tdp.TextDocument.URI) {
		return "", errIgnored
	}

	return tdp.TextDocument.URI, nil
}

func decodeParams[P any](req *jsonrpc2.Request) (P, error) {
	var params P

	err := handler.Decode(req, &params)

	return params, err
}
