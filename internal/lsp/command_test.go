package lsp

import (
	"encoding/json"
	"testing"

	"github.com/styrainc/regal/internal/lsp/types"
)

func TestUseAssignmentOperatorCommand(t *testing.T) {
	t.Parallel()

	target := "file:///example.rego"
	diag := types.Diagnostic{
		Code:    "use-assignment-operator",
		Message: "Replace = with := in assignment",
	}

	expectedArgs, err := json.Marshal(commandArgs{Target: target, Diagnostic: &diag})
	if err != nil {
		t.Fatalf("Unexpected error marshalling command arguments: %s", err)
	}

	expectedCommand := types.Command{
		Title:     "Replace = with := in assignment",
		Command:   "regal.fix.use-assignment-operator",
		Tooltip:   "Replace = with := in assignment",
		Arguments: toAnySlice(string(expectedArgs)),
	}

	result := UseAssignmentOperatorCommand(target, diag)

	if exp, got := expectedCommand.Title, result.Title; exp != got {
		t.Fatalf("Expected Title %q, got %q", exp, got)
	}

	if exp, got := expectedCommand.Command, result.Command; exp != got {
		t.Fatalf("Expected Command %q, got %q", exp, got)
	}

	if exp, got := expectedCommand.Tooltip, result.Tooltip; exp != got {
		t.Fatalf("Expected Tooltip %q, got %q", exp, got)
	}

	if exp, got := len(*expectedCommand.Arguments), len(*result.Arguments); exp != got {
		t.Fatalf("Expected %d arguments, got %d", exp, got)
	}

	if exp, got := (*expectedCommand.Arguments)[0], (*result.Arguments)[0]; exp != got {
		t.Fatalf("Expected argument \n%s\n%s", exp, got)
	}
}
