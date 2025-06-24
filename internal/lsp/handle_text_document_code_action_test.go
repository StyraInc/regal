package lsp

import (
	"reflect"
	"testing"

	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/internal/web"

	"github.com/styrainc/roast/pkg/encoding"
)

func TestHandleTextDocumentCodeAction(t *testing.T) {
	t.Parallel()

	webServer := &web.Server{}
	webServer.SetBaseURL("http://foo.bar")

	l := &LanguageServer{
		clientIdentifier: clients.IdentifierGeneric,
		webServer:        webServer,
	}

	uri := "file:///example.rego"

	diag := types.Diagnostic{
		Code:    ruleNameUseAssignmentOperator,
		Message: "foobar",
		Range: types.Range{
			Start: types.Position{Line: 2, Character: 4},
			End:   types.Position{Line: 2, Character: 10},
		},
	}

	params := types.CodeActionParams{
		TextDocument: types.TextDocumentIdentifier{URI: uri},
		Context:      types.CodeActionContext{Diagnostics: []types.Diagnostic{diag}},
		Range: types.Range{
			Start: types.Position{Line: 2, Character: 4},
			End:   types.Position{Line: 2, Character: 10},
		},
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
				Target:     uri,
				Diagnostic: &diag,
			})))),
		},
	}

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

	actualAction := actions[0]

	if exp, got := expectedAction.Title, actualAction.Title; exp != got {
		t.Fatalf("expected Title %q, got %q", exp, got)
	}

	if exp, got := expectedAction.Kind, actualAction.Kind; exp != got {
		t.Fatalf("expected Kind %q, got %q", exp, got)
	}

	if exp, got := len(expectedAction.Diagnostics), len(actualAction.Diagnostics); exp != got {
		t.Fatalf("expected %d diagnostics, got %d", exp, got)
	}

	if exp, got := *expectedAction.IsPreferred, *actualAction.IsPreferred; actualAction.IsPreferred == nil || exp != got {
		t.Fatalf("expected IsPreferred to be %v, got %v", exp, actualAction.IsPreferred)
	}

	if exp, got := expectedAction.Command.Command, actualAction.Command.Command; exp != got {
		t.Fatalf("expected Command %q, got %q", exp, got)
	}

	if actualAction.Command.Arguments == nil {
		t.Fatal("expected Command.Arguments to be non-nil")
	}

	expArgs, actualArgs := *expectedAction.Command.Arguments, *actualAction.Command.Arguments
	if exp, got := len(expArgs), len(actualArgs); exp != got {
		t.Fatalf("expected %d arguments, got %d", exp, got)
	}

	var expDecoded, actualDecoded map[string]any

	if err = encoding.JSON().Unmarshal([]byte(expArgs[0].(string)), &expDecoded); err != nil {
		t.Fatalf("failed to unmarshal expected arguments: %v", err)
	}

	if err = encoding.JSON().Unmarshal([]byte(actualArgs[0].(string)), &actualDecoded); err != nil {
		t.Fatalf("failed to unmarshal actual arguments: %v", err)
	}

	if !reflect.DeepEqual(expDecoded, actualDecoded) {
		t.Errorf("expected Command.Arguments to be %v, got %v", expDecoded, actualDecoded)
	}
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
			Diagnostics: []types.Diagnostic{
				{
					Code:    ruleNameUseAssignmentOperator,
					Message: "foobar",
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 4},
						End:   types.Position{Line: 2, Character: 10},
					},
				},
			},
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
