package lsp

import (
	"testing"
)

func TestTokenFoldingRanges(t *testing.T) {
	t.Parallel()

	policy := `package p

	import rego.v1

rule if {
	arr := [
		1,
		2,
		3,
	]
	par := (
		1 +
		2 -
		3
	)
}`

	foldingRanges := TokenFoldingRanges(policy)

	if len(foldingRanges) != 3 {
		t.Fatalf("Expected 3 folding ranges, got %d", len(foldingRanges))
	}

	arr := foldingRanges[0]

	if arr.StartLine != 5 || *arr.StartCharacter != 9 {
		t.Errorf("Expected start line 5 and start character 9, got %d and %d", arr.StartLine, *arr.StartCharacter)
	}

	if arr.EndLine != 8 {
		t.Errorf("Expected end line 8, got %d", arr.EndLine)
	}

	parens := foldingRanges[1]

	if parens.StartLine != 10 || *parens.StartCharacter != 9 {
		t.Errorf("Expected start line 10 and start character 9, got %d and %d", parens.StartLine, *parens.StartCharacter)
	}

	if parens.EndLine != 13 {
		t.Errorf("Expected end line 13, got %d", parens.EndLine)
	}

	rule := foldingRanges[2]

	if rule.StartLine != 4 || *rule.StartCharacter != 9 {
		t.Errorf("Expected start line 4 and start character 9, got %d and %d", rule.StartLine, *rule.StartCharacter)
	}

	if rule.EndLine != 14 {
		t.Errorf("Expected end line 7, got %d", rule.EndLine)
	}
}

func TestTokenInvalidFoldingRanges(t *testing.T) {
	t.Parallel()

	policy := `package p

arr := ]]`

	foldingRanges := TokenFoldingRanges(policy)

	if len(foldingRanges) != 0 {
		t.Fatalf("Expected no folding ranges, got %d", len(foldingRanges))
	}
}
