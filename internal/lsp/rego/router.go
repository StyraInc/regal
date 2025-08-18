package rego

import (
	"context"
	"errors"
	"strings"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/open-policy-agent/opa/v1/storage"

	"github.com/styrainc/regal/internal/lsp/handler"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/util"
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
		ContextProvider              func(uri string, reqs *Requirements) RegalContext
		ContentProvider              func(uri string) (string, bool)
		IgnoredProvider              func(uri string) bool
		ParseErrorsProvider          func(uri string) ([]types.Diagnostic, bool)
		SuccessfulParseCountProvider func(uri string) (int, bool)
	}

	RegoRouter struct {
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

func NewRegoRouter(ctx context.Context, store storage.Store, prvs Providers) *RegoRouter {
	if err := StoreCachedQuery(ctx, entrypoint, store); err != nil {
		panic(err)
	}

	routes := map[string]Route{
		"textDocument/codeAction": {
			handler: textDocument[types.CodeActionParams, []types.CodeAction],
		},
		"textDocument/codeLens": {
			handler: textDocument[types.CodeLensParams, []types.CodeLens],
			requires: &Requirements{
				File: FileRequirements{
					Lines:                    true,
					SuccessfulParseLineCount: true,
					ParseErrors:              true,
				},
			},
		},
		"textDocument/documentLink": {
			handler: textDocument[types.DocumentLinkParams, []types.DocumentLink],
		},
		"textDocument/documentHighlight": {
			handler:  textDocument[types.DocumentHighlightParams, []types.DocumentHighlight],
			requires: &Requirements{File: FileRequirements{Lines: true}},
		},
		"textDocument/signatureHelp": {
			handler:  textDocument[types.SignatureHelpParams, *types.SignatureHelp],
			requires: &Requirements{File: FileRequirements{Lines: true}},
		},
	}

	return &RegoRouter{routes: routes, providers: prvs}
}

func (m *RegoRouter) Handle(ctx context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
	if route, ok := m.routes[req.Method]; ok {
		return requirementsHandler(route)(ctx, m.providers, req)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: "method not supported: " + req.Method}
}

// requirementsHandler wraps a regoHandler which first verifies that the text document URI isn't ignored
// and then goes on to ensure that any custom requirements the handler may have are met.
func requirementsHandler(route Route) regoHandler {
	return func(ctx context.Context, prvs Providers, req *jsonrpc2.Request) (any, error) {
		// This is mandatory requirement for all routes managed here.
		uri, err := decodeAndCheckURI(req, prvs.IgnoredProvider)
		if err != nil {
			if errors.Is(err, errIgnored) {
				return emptyResponse[req.Method], nil
			}

			return nil, err
		}

		// Set up a basic RegalContext, which while not used by all routes, is provided for all.
		rctx := prvs.ContextProvider(uri, route.requires)

		if route.requires == nil {
			return route.handler(ctx, rctx, req)
		}

		if route.requires.File.Lines && rctx.File.Lines == nil {
			if prvs.ContentProvider == nil {
				return nil, errors.New("content provider required but not provided")
			}

			content, ok := prvs.ContentProvider(uri)
			if !ok {
				return nil, errors.New("content provider failed to provide content for URI: " + uri)
			}

			rctx.File.Lines = strings.Split(content, "\n")
		}

		if route.requires.File.SuccessfulParseLineCount {
			if prvs.SuccessfulParseCountProvider == nil {
				return nil, errors.New("successful parse count provider required but not provided")
			}

			if splc, ok := prvs.SuccessfulParseCountProvider(uri); ok {
				rctx.File.SuccessfulParseCount = util.SafeIntToUint(splc)
			} else {
				// if the file has always been unparsable, we can return early
				return emptyResponse[req.Method], nil
			}
		}

		if route.requires.File.ParseErrors {
			if prvs.ParseErrorsProvider == nil {
				return nil, errors.New("parse errors provider required but not provided")
			}

			if rctx.File.ParseErrors, _ = prvs.ParseErrorsProvider(uri); rctx.File.ParseErrors == nil {
				rctx.File.ParseErrors = make([]types.Diagnostic, 0)
			}
		}

		return route.handler(ctx, rctx, req)
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

	if ignored != nil && ignored(tdp.TextDocument.URI) {
		return "", errIgnored
	}

	return tdp.TextDocument.URI, nil
}

func decodeParams[P any](req *jsonrpc2.Request) (P, error) {
	var params P

	err := handler.Decode(req, &params)

	return params, err
}
