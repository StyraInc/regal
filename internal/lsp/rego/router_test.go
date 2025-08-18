package rego_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/storage/inmem"

	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/rego"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/testutil"
	"github.com/styrainc/regal/pkg/roast/encoding"
)

func TestRouteTextDocumentCodeAction(t *testing.T) {
	t.Parallel()

	mgr := rego.NewRegoRouter(t.Context(), nil, providers(regalContext(), "", ""))
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

	mgr := rego.NewRegoRouter(t.Context(), store, providers(regalContext(), "", ""))
	req := &jsonrpc2.Request{Method: "textDocument/documentLink", Params: linkParams(t, fileURI)}

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

func TestRouteTextDocumentDocumentHighlight(t *testing.T) {
	t.Parallel()

	module := parse.MustParseModule("# METADATA\n# title: p\npackage p\n")
	moduleMap := make(map[string]any)
	fileURI := "file:///workspace/p.rego"

	encoding.MustJSONRoundTrip(module, &moduleMap)

	store := inmem.NewFromObjectWithOpts(map[string]any{
		"workspace": map[string]any{
			"parsed": map[string]any{
				fileURI: moduleMap,
			},
		},
	}, inmem.OptRoundTripOnWrite(false))

	mgr := rego.NewRegoRouter(t.Context(), store, rego.Providers{
		ContextProvider: func(string, *rego.Requirements) rego.RegalContext {
			return regalContext()
		},
		ContentProvider: func(uri string) (string, bool) {
			if uri == fileURI {
				return "# METADATA\n# title: p\npackage p\n", true
			}

			return "", false
		},
	})

	req := &jsonrpc2.Request{
		Method: "textDocument/documentHighlight",
		Params: docPositionParams(t, fileURI, types.Position{Line: 0, Character: 4}),
	}
	resp := testutil.Must(mgr.Handle(t.Context(), nil, req))(t)

	if resp == nil {
		t.Fatal("expected a response, got nil")
	}

	result, ok := resp.([]types.DocumentHighlight)
	if !ok {
		t.Fatalf("expected response to be of type rego.Result[[]types.DocumentLink], got %T", resp)
	}

	if len(result) == 0 {
		t.Fatal("expected at least one document link, got none")
	}
}

func TestRouteIgnoredDocument(t *testing.T) {
	t.Parallel()

	mgr := rego.NewRegoRouter(t.Context(), nil, providers(regalContext(), "", "file:///workspace/ignored.rego"))
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

func TestTextDocumentSignatureHelp(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	content := `package example

allow if regex.match(` + "`foo`" + `, "bar")
allow if count([1,2,3]) == 2
allow if concat(",", "a", "b") == "b,a"`

	module := parse.MustParseModule(content)
	moduleMap := make(map[string]any)

	encoding.MustJSONRoundTrip(module, &moduleMap)

	store := inmem.NewFromObjectWithOpts(map[string]any{
		"workspace": map[string]any{
			"builtins": map[string]any{
				"count":       ast.Count,
				"concat":      ast.Concat,
				"regex.match": ast.RegexMatch,
			},
			"parsed": map[string]any{
				"file://workspace/p.rego": moduleMap,
			},
		},
	}, inmem.OptRoundTripOnWrite(true))

	testCases := map[string]struct {
		position       types.Position
		expectedLabel  string
		expectedDoc    string
		expectedParams []string
	}{
		"regex.match function call": {
			position:       types.Position{Line: 2, Character: 21},
			expectedLabel:  "regex.match(pattern: string, value: string) -> boolean",
			expectedDoc:    "Matches a string against a regular expression.",
			expectedParams: []string{"pattern: string", "value: string"},
		},
		"count function call": {
			position:       types.Position{Line: 3, Character: 16},
			expectedLabel:  "count(collection: any) -> number",
			expectedDoc:    "Count takes a collection or string and returns the number of elements (or characters) in it.",
			expectedParams: []string{"collection: any"},
		},
		"concat function call": {
			position:       types.Position{Line: 4, Character: 17},
			expectedLabel:  "concat(delimiter: string, collection: any) -> string",
			expectedDoc:    "Joins a set or array of strings with a delimiter.",
			expectedParams: []string{"delimiter: string", "collection: any"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mgr := rego.NewRegoRouter(ctx, store, providers(regalContext(), content, ""))
			req := &jsonrpc2.Request{
				Method: "textDocument/signatureHelp",
				Params: docPositionParams(t, "file://workspace/p.rego", tc.position),
			}

			resp, err := mgr.Handle(ctx, nil, req)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if resp == nil {
				t.Fatal("expected a response, got nil")
			}

			signatureHelp, ok := resp.(*types.SignatureHelp)
			if !ok {
				t.Fatalf("expected response to be of type types.SignatureHelp, got %T", resp)
			}

			if signatureHelp == nil {
				t.Fatalf("signature help nil at position line=%d character=%d", tc.position.Line, tc.position.Character)
			}

			if len(signatureHelp.Signatures) == 0 {
				t.Error("expected at least one signature")
			}

			if signatureHelp.ActiveSignature == nil {
				t.Error("expected ActiveSignature to be set")
			} else if *signatureHelp.ActiveSignature != 0 {
				t.Errorf("expected ActiveSignature to be 0, got %d", *signatureHelp.ActiveSignature)
			}

			if signatureHelp.ActiveParameter == nil {
				t.Error("expected ActiveParameter to be set")
			} else if *signatureHelp.ActiveParameter != 0 {
				t.Errorf("expected ActiveParameter to be 0, got %d", *signatureHelp.ActiveParameter)
			}

			sig := signatureHelp.Signatures[0]

			if sig.Label != tc.expectedLabel {
				t.Errorf("expected signature label to be '%s', got '%s'", tc.expectedLabel, sig.Label)
			}

			if sig.Documentation != tc.expectedDoc {
				t.Errorf("expected documentation to be '%s', got '%s'", tc.expectedDoc, sig.Documentation)
			}

			if len(sig.Parameters) != len(tc.expectedParams) {
				t.Fatalf("expected %d parameters, got %d", len(tc.expectedParams), len(sig.Parameters))
			}

			for i, expectedParam := range tc.expectedParams {
				if sig.Parameters[i].Label != expectedParam {
					t.Errorf("expected parameter %d label to be '%s', got '%s'", i, expectedParam, sig.Parameters[i].Label)
				}
			}

			if sig.ActiveParameter == nil {
				t.Error("expected signature ActiveParameter to be set")
			} else if *sig.ActiveParameter != 0 {
				t.Errorf("expected signature ActiveParameter to be 0, got %d", *sig.ActiveParameter)
			}
		})
	}
}

func docPositionParams(t *testing.T, uri string, position types.Position) *json.RawMessage {
	t.Helper()

	return testutil.ToJSONRawMessage(t, map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"position": position,
	})
}

func codeActionParams(t *testing.T, uri string, ls, cs, le, ce int) *json.RawMessage {
	t.Helper()

	return testutil.ToJSONRawMessage(t, map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"range": map[string]any{
			"start": map[string]int{"line": ls, "character": cs},
			"end":   map[string]int{"line": le, "character": ce},
		},
	})
}

func linkParams(t *testing.T, uri string) *json.RawMessage {
	t.Helper()

	return testutil.ToJSONRawMessage(t, map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
	})
}

func providers(rc rego.RegalContext, content, ignored string) rego.Providers {
	return rego.Providers{
		ContextProvider: func(string, *rego.Requirements) rego.RegalContext {
			return rc
		},
		IgnoredProvider: func(uri string) bool {
			return uri == ignored
		},
		ContentProvider: func(_ string) (string, bool) {
			if content != "" {
				return content, true
			}

			return "", false
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
		File: rego.File{
			Name: "workspace/p.rego",
			Abs:  "/workspace/p.rego",
		},
	}
}
