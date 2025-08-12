package providers

import (
	"slices"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/rego"
	"github.com/styrainc/regal/internal/lsp/test"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/testutil"
)

var opts = &Options{Builtins: rego.BuiltinsForCapabilities(ast.CapabilitiesForThisVersion())}

func TestBuiltIns_if(t *testing.T) {
	t.Parallel()

	c, p := setupCacheAndProvider(t, "package foo\n\nallow if c")
	completions := testutil.Must(p.Run(t.Context(), c, types.NewCompletionParams(testCaseFileURI, 2, 10, nil), opts))(t)

	if labels := test.Labels(completions); !slices.Contains(labels, "count") {
		t.Fatalf("Expected to find 'count' in completions, got: %s", strings.Join(labels, ", "))
	}
}

func TestBuiltIns_afterAssignment(t *testing.T) {
	t.Parallel()

	c, p := setupCacheAndProvider(t, "package foo\n\nallow := c")
	completions := testutil.Must(p.Run(t.Context(), c, types.NewCompletionParams(testCaseFileURI, 2, 10, nil), opts))(t)

	if labels := test.Labels(completions); !slices.Contains(labels, "count") {
		t.Fatalf("Expected to find 'count' in completions, got: %s", strings.Join(labels, ", "))
	}
}

func TestBuiltIns_inRuleBody(t *testing.T) {
	t.Parallel()

	c, p := setupCacheAndProvider(t, "package foo\n\nallow if {\n  c\n}")
	completions := testutil.Must(p.Run(t.Context(), c, types.NewCompletionParams(testCaseFileURI, 3, 3, nil), opts))(t)

	if labels := test.Labels(completions); !slices.Contains(labels, "count") {
		t.Fatalf("Expected to find 'count' in completions, got: %s", strings.Join(labels, ", "))
	}
}

func TestBuiltIns_noInfix(t *testing.T) {
	t.Parallel()

	c, p := setupCacheAndProvider(t, "package foo\n\nallow if gt")

	completions := testutil.Must(p.Run(t.Context(), c, types.NewCompletionParams(testCaseFileURI, 2, 10, nil), opts))(t)
	if len(completions) != 0 {
		t.Fatalf("Expected no completions, got: %v", completions)
	}
}

func TestBuiltIns_noDeprecated(t *testing.T) {
	t.Parallel()

	c, p := setupCacheAndProvider(t, "package foo\n\nallow if c")
	completions := testutil.Must(p.Run(t.Context(), c, types.NewCompletionParams(testCaseFileURI, 2, 10, nil), opts))(t)

	if labels := test.Labels(completions); slices.Contains(labels, "cast_set") {
		t.Fatalf("Expected no deprecated completions, got: %s", strings.Join(labels, ", "))
	}
}

func TestBuiltIns_noDefaultRule(t *testing.T) {
	t.Parallel()

	c, p := setupCacheAndProvider(t, "package foo\n\ndefault allow := f")

	completions := testutil.Must(p.Run(t.Context(), c, types.NewCompletionParams(testCaseFileURI, 2, 18, nil), opts))(t)
	if len(completions) != 0 {
		t.Fatalf("Expected no completions, got: %d", len(completions))
	}
}

func setupCacheAndProvider(t *testing.T, contents string) (*cache.Cache, *BuiltIns) {
	t.Helper()

	c := cache.NewCache()
	c.SetFileContents(testCaseFileURI, contents)

	return c, &BuiltIns{}
}
