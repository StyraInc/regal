package rego

import (
	"context"
	"testing"

	"github.com/styrainc/regal/internal/parse"
)

func TestCodeLenses(t *testing.T) {
	t.Parallel()

	contents := `package p

	import rego.v1

	allow if "foo" in input.bar`

	module := parse.MustParseModule(contents)

	lenses, er := CodeLenses(context.TODO(), "p.rego", contents, module)
	if er != nil {
		t.Fatalf("unexpected error: %v", er)
	}

	// 2 for the package, 2 for the rule
	// the contents of the lenses are tested in Rego
	if len(lenses) != 4 {
		t.Fatalf("expected 4 code lenses, got %d", len(lenses))
	}
}

func TestAllRuleHeadLocations(t *testing.T) {
	t.Parallel()

	contents := `package p

	import rego.v1

	default allow := false

	allow if 1
	allow if 2

	foo.bar[x] if x := 1
	foo.bar[x] if x := 2`

	module := parse.MustParseModule(contents)

	ruleHeads, er := AllRuleHeadLocations(context.TODO(), "p.rego", contents, module)
	if er != nil {
		t.Fatalf("unexpected error: %v", er)
	}

	if len(ruleHeads) != 2 {
		t.Fatalf("expected 2 code lenses, got %d", len(ruleHeads))
	}

	if len(ruleHeads["data.p.allow"]) != 3 {
		t.Fatalf("expected 3 allow rule heads, got %d", len(ruleHeads["data.p.allow"]))
	}

	if len(ruleHeads["data.p.foo.bar"]) != 2 {
		t.Fatalf("expected 2 foo.bar rule heads, got %d", len(ruleHeads["data.p.foo.bar"]))
	}
}

func TestAllKeywords(t *testing.T) {
	t.Parallel()

	contents := `package p

	import rego.v1

	my_set contains "x" if true
	`

	module := parse.MustParseModule(contents)

	keywords, er := AllKeywords(context.TODO(), "p.rego", contents, module)
	if er != nil {
		t.Fatalf("unexpected error: %v", er)
	}

	// this is "lines with keywords", not number of keywords
	if len(keywords) != 3 {
		t.Fatalf("expected 1 keyword, got %d", len(keywords))
	}

	if len(keywords["1"]) != 1 {
		t.Fatalf("expected 1 keywords on line 1, got %d", len(keywords["1"]))
	}

	if len(keywords["3"]) != 1 {
		t.Fatalf("expected 1 keywords on line 3, got %d", len(keywords["1"]))
	}

	if len(keywords["5"]) != 2 {
		t.Fatalf("expected 2 keywords on line 5, got %d", len(keywords["1"]))
	}
}
