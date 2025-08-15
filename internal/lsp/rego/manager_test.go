package rego_test

import (
	"encoding/json"
	"testing"

	"github.com/open-policy-agent/opa/v1/storage/inmem"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/rego"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/testutil"
	"github.com/styrainc/regal/pkg/roast/encoding"
)

func TestRouteTextDocumentCodeAction(t *testing.T) {
	mgr := rego.NewRegoManager(t.Context(), nil, providers(regalContext(), ""))
	req := &jsonrpc2.Request{
		Method: "textDocument/codeAction",
		Params: codeActionParams(t, "file:///workspace/p.rego", 0, 0, 0, 10),
	}

	resp, err := mgr.Handle(t.Context(), nil, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("expected a response, got nil")
	}

	result, ok := resp.([]types.CodeAction)
	if !ok {
		t.Fatalf("expected response to be of type []types.CodeAction, got %T", resp)
	}

	if len(result) == 0 {
		t.Fatal("expected at least one code action, got none")
	}
}

func TestRouteTextDocumentDocumentLink(t *testing.T) {
	t.Parallel()

	module := parse.MustParseModule("# regal ignore:prefer-snake-case\npackage p\n")
	moduleMap := make(map[string]any)
	fileURI := "file:///workspace/p.rego"

	encoding.MustJSONRoundTrip(module, &moduleMap)

	store := inmem.NewFromObjectWithOpts(map[string]any{
		"workspace": map[string]any{
			"parsed": map[string]any{
				fileURI: moduleMap,
			},
			"config": map[string]any{
				"rules": map[string]any{
					"style": map[string]any{
						"prefer-snake-case": map[string]any{},
					},
				},
			},
		},
	}, inmem.OptRoundTripOnWrite(false))

	mgr := rego.NewRegoManager(t.Context(), store, providers(regalContext(), ""))
	req := &jsonrpc2.Request{Method: "textDocument/documentLink", Params: documentLinkParams(t, fileURI)}

	resp := testutil.Must(mgr.Handle(t.Context(), nil, req))(t)

	if resp == nil {
		t.Fatal("expected a response, got nil")
	}

	result, ok := resp.([]types.DocumentLink)
	if !ok {
		t.Fatalf("expected response to be of type rego.Result[[]types.DocumentLink], got %T", resp)
	}

	if len(result) == 0 {
		t.Fatal("expected at least one document link, got none")
	}
}

func TestRouteIgnoredDocument(t *testing.T) {
	mgr := rego.NewRegoManager(t.Context(), nil, providers(regalContext(), "file:///workspace/ignored.rego"))
	req := &jsonrpc2.Request{
		Method: "textDocument/codeAction",
		Params: codeActionParams(t, "file:///workspace/ignored.rego", 0, 0, 0, 10),
	}

	resp := testutil.Must(mgr.Handle(t.Context(), nil, req))(t)

	if resp == nil {
		t.Fatal("expected a response, got nil")
	}

	result, ok := resp.([]types.CodeAction)
	if !ok {
		t.Fatalf("expected response to be of type []types.CodeAction, got %T", resp)
	}

	if result == nil {
		t.Fatal("expected an empty response, got nil")
	}

	if len(result) != 0 {
		t.Fatal("expected empty response, got code actions")
	}
}

func codeActionParams(t *testing.T, uri string, ls, cs, le, ce int) *json.RawMessage {
	t.Helper()

	return testutil.ToJsonRawMessage(t, map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"range": map[string]any{
			"start": map[string]int{"line": ls, "character": cs},
			"end":   map[string]int{"line": le, "character": ce},
		},
	})
}

func documentLinkParams(t *testing.T, uri string) *json.RawMessage {
	t.Helper()

	return testutil.ToJsonRawMessage(t, map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
	})
}

func providers(rc rego.RegalContext, ignored string) rego.Providers {
	return rego.Providers{
		ContextProvider: func(uri string, reqs *rego.Requirements) rego.RegalContext {
			return rc
		},
		IgnoredProvider: func(uri string) bool {
			return uri == ignored
		},
	}
}

func regalContext() rego.RegalContext {
	return rego.RegalContext{
		Client: types.Client{
			Identifier: clients.IdentifierVSCode,
		},
		Environment: rego.Environment{
			PathSeparator:    "/",
			WorkspaceRootURI: "file:///workspace",
			WebServerBaseURI: "http://webserver",
		},
	}
}
