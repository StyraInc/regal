package lsp

import (
	"reflect"
	"testing"

	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/internal/web"
	"github.com/styrainc/regal/pkg/roast/encoding"
)

func TestHandleTextDocumentCodeAction(t *testing.T) {
	t.Parallel()

	webServer := &web.Server{}
	webServer.SetBaseURL("http://foo.bar")

	l := &LanguageServer{clientIdentifier: clients.IdentifierGeneric, webServer: webServer}

	diag := types.Diagnostic{
		Code:    ruleNameUseAssignmentOperator,
		Message: "foobar",
		Range:   types.RangeBetween(2, 4, 2, 10),
	}

	params := types.CodeActionParams{
		TextDocument: types.TextDocumentIdentifier{URI: "file:///example.rego"},
		Context:      types.CodeActionContext{Diagnostics: []types.Diagnostic{diag}},
		Range:        types.RangeBetween(2, 4, 2, 10),
	}

	expectedAction := types.CodeAction{
		Title:       "Replace = with := in assignment",
		Kind:        "quickfix",
		Diagnostics: params.Context.Diagnostics,
		IsPreferred: truePtr,
		Command: types.Command{
			Title:   "Replace = with := in assignment",
			Command: "regal.fix.use-assignment-operator",
			Tooltip: "Replace = with := in assignment",
			Arguments: toAnySlice(string(util.Must(encoding.JSON().Marshal(commandArgs{
				Target:     params.TextDocument.URI,
				Diagnostic: &diag,
			})))),
		},
	}

	actualAction := invokeCodeActionHandler(t, l, params)

	assertExpectedCodeAction(t, expectedAction, actualAction)

	expArgs, actualArgs := *expectedAction.Command.Arguments, *actualAction.Command.Arguments
	if exp, got := len(expArgs), len(actualArgs); exp != got {
		t.Fatalf("expected %d arguments, got %d", exp, got)
	}

	var expDecoded, actualDecoded map[string]any

	if err := encoding.JSON().Unmarshal([]byte(expArgs[0].(string)), &expDecoded); err != nil {
		t.Fatalf("failed to unmarshal expected arguments: %v", err)
	}

	if err := encoding.JSON().Unmarshal([]byte(actualArgs[0].(string)), &actualDecoded); err != nil {
		t.Fatalf("failed to unmarshal actual arguments: %v", err)
	}

	if !reflect.DeepEqual(expDecoded, actualDecoded) {
		t.Errorf("expected Command.Arguments to be %v, got %v", expDecoded, actualDecoded)
	}
}

func TestHandleTextDocumentCodeActionSourceExplorer(t *testing.T) {
	t.Parallel()

	webServer := &web.Server{}
	webServer.SetBaseURL("http://foo.bar")

	l := &LanguageServer{
		clientIdentifier:            clients.IdentifierVSCode,
		clientInitializationOptions: types.InitializationOptions{},
		webServer:                   webServer,
		workspaceRootURI:            "file:///foo",
	}

	params := types.CodeActionParams{
		TextDocument: types.TextDocumentIdentifier{URI: "file:///foo/example.rego"},
		Context:      types.CodeActionContext{},
		Range:        types.RangeBetween(2, 4, 2, 10),
	}

	expectedAction := types.CodeAction{
		Title: "Explore compiler stages for this policy",
		Kind:  "source.explore",
		Command: types.Command{
			Title:     "Explore compiler stages for this policy",
			Command:   "vscode.open",
			Tooltip:   "Explore compiler stages for this policy",
			Arguments: toAnySlice("http://foo.bar/explorer/example.rego"),
		},
	}

	actualAction := invokeCodeActionHandler(t, l, params)

	assertExpectedCodeAction(t, expectedAction, actualAction)

	expArgs, actualArgs := *expectedAction.Command.Arguments, *actualAction.Command.Arguments
	if exp, got := len(expArgs), len(actualArgs); exp != got {
		t.Fatalf("expected %d arguments, got %d", exp, got)
	}
}

func assertExpectedCodeAction(t *testing.T, expected, actual types.CodeAction) {
	t.Helper()

	if expected.Title != actual.Title {
		t.Errorf("expected Title %q, got %q", expected.Title, actual.Title)
	}

	if expected.Kind != actual.Kind {
		t.Errorf("expected Kind %q, got %q", expected.Kind, actual.Kind)
	}

	if len(expected.Diagnostics) != len(actual.Diagnostics) {
		t.Errorf("expected %d diagnostics, got %d", len(expected.Diagnostics), len(actual.Diagnostics))
	}

	if expected.IsPreferred == nil && actual.IsPreferred != nil { //nolint:gocritic
		t.Error("expected IsPreferred to be nil")
	} else if expected.IsPreferred != nil && actual.IsPreferred == nil {
		t.Error("expected IsPreferred to be non-nil")
	} else if expected.IsPreferred != nil && actual.IsPreferred != nil && *expected.IsPreferred != *actual.IsPreferred {
		t.Errorf("expected IsPreferred to be %v, got %v", *expected.IsPreferred, *actual.IsPreferred)
	}

	if expected.Command.Command != actual.Command.Command {
		t.Errorf("expected Command %q, got %q", expected.Command.Command, actual.Command.Command)
	}

	if expected.Command.Title != actual.Command.Title {
		t.Errorf("expected Command.Title %q, got %q", expected.Command.Title, actual.Command.Title)
	}

	if expected.Command.Tooltip != actual.Command.Tooltip {
		t.Errorf("expected Command.Tooltip %q, got %q", expected.Command.Tooltip, actual.Command.Tooltip)
	}

	// Just check nilness here, and leave the actual content to the test.
	if expected.Command.Arguments == nil && actual.Command.Arguments != nil {
		t.Error("expected Command.Arguments to be nil")
	} else if expected.Command.Arguments != nil && actual.Command.Arguments == nil {
		t.Error("expected Command.Arguments to be non-nil")
	}
}

func invokeCodeActionHandler(t *testing.T, l *LanguageServer, params types.CodeActionParams) types.CodeAction {
	t.Helper()

	result, err := l.handleTextDocumentCodeAction(t.Context(), params)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	actions, ok := result.([]types.CodeAction)
	if !ok {
		t.Errorf("Expected result to be of type []types.CodeAction, got %T", result)
	}

	if exp, got := 1, len(actions); exp != got {
		t.Fatalf("Expected %d action, got %d", exp, got)
	}

	return actions[0]
}

// 63243 ns/op	   59576 B/op	    1110 allocs/op - the OPA JSON roundtrip method
// 42402 ns/op	   37822 B/op	     738 allocs/op - build input Value by hand
// 45049 ns/op	   39731 B/op	     790 allocs/op - build input Value using reflection
// 44024 ns/op	   38040 B/op	     749 allocs/op - build input Value using reflection + interning
// ...
// "real world" usage shows a number somewhere between 0.1 - 0.5 ms
// of which most of the cost is in JSON marshaling and unmarshaling.
func BenchmarkHandleTextDocumentCodeAction(b *testing.B) {
	l := &LanguageServer{
		clientIdentifier: clients.IdentifierGeneric,
		webServer:        &web.Server{},
	}

	params := types.CodeActionParams{
		TextDocument: types.TextDocumentIdentifier{URI: "file:///example.rego"},
		Context: types.CodeActionContext{
			Diagnostics: []types.Diagnostic{{
				Code:    ruleNameUseAssignmentOperator,
				Message: "foobar",
				Range:   types.RangeBetween(2, 4, 2, 10),
			}},
		},
	}

	for b.Loop() {
		res, err := l.handleTextDocumentCodeAction(b.Context(), params)
		if err != nil {
			b.Fatal(err)
		}

		if len(res.([]types.CodeAction)) != 1 {
			b.Fatalf("expected 1 code action, got %d", len(res.([]types.CodeAction)))
		}
	}
}

func toAnySlice(a ...string) *[]any {
	b := make([]any, len(a))
	for i := range a {
		b[i] = a[i]
	}

	return &b
}
