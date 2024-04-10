package rules_test

import (
	"context"
	"testing"

	"github.com/styrainc/regal/internal/test"
	"github.com/styrainc/regal/internal/testutil"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/rules"
)

func TestFmtRuleFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	policy := "  package p "

	result := testutil.Must(rules.NewOpaFmtRule(config.Config{}).Run(ctx, test.InputPolicy("p.rego", policy)))(t)

	if len(result.Violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "opa-fmt" {
		t.Errorf("expected violation title to be 'opa-fmt', got %s", result.Violations[0].Title)
	}

	if result.Violations[0].Category != "style" {
		t.Errorf("expected violation category to be 'style', got %s", result.Violations[0].Category)
	}

	if result.Violations[0].Location.File != "p.rego" {
		t.Errorf("expected violation location file to be 'p.rego', got %s", result.Violations[0].Location.File)
	}

	if result.Violations[0].Location.Row != 1 {
		t.Errorf("expected violation location row to be 1, got %d", result.Violations[0].Location.Row)
	}

	if result.Violations[0].Location.Column != 1 {
		t.Errorf("expected violation location column to be 1, got %d", result.Violations[0].Location.Column)
	}

	if *result.Violations[0].Location.Text != "  package p " {
		t.Errorf("expected violation location text to be '  package p ', got %q", *result.Violations[0].Location.Text)
	}
}

func TestFmtRuleSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	policy := "package p\n"

	result := testutil.Must(rules.NewOpaFmtRule(config.Config{}).Run(ctx, test.InputPolicy("p.rego", policy)))(t)

	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violation, got %d", len(result.Violations))
	}
}
