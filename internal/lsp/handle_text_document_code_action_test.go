package lsp

import (
	"slices"
	"testing"

	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
)

func TestHandleTextDocumentCodeAction(t *testing.T) {
	t.Parallel()

	l := &LanguageServer{clientIdentifier: clients.IdentifierGeneric}

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
		TextDocument: types.TextDocumentIdentifier{
			URI: uri,
		},
		Context: types.CodeActionContext{
			Diagnostics: []types.Diagnostic{diag},
		},
	}

	expectedAction := types.CodeAction{
		Title:       "Replace = with := in assignment",
		Kind:        "quickfix",
		Diagnostics: params.Context.Diagnostics,
		IsPreferred: truePtr,
		Command:     UseAssignmentOperatorCommand(uri, diag),
	}

	result, err := l.handleTextDocumentCodeAction(params)
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

	if exp, got := *expectedAction.Command.Arguments, *actualAction.Command.Arguments; !slices.Equal(exp, got) {
		t.Fatalf("expected Arguments %v, got %v", exp, got)
	}
}
