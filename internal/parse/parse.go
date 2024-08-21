package parse

import (
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"

	rio "github.com/styrainc/regal/internal/io"
)

// ParserOptions provides the parse options necessary to include location data in AST results.
func ParserOptions() ast.ParserOptions {
	return ast.ParserOptions{
		ProcessAnnotation: true,
	}
}

// MustParseModule works like ast.MustParseModule but with the Regal parser options applied.
func MustParseModule(policy string) *ast.Module {
	return ast.MustParseModuleWithOpts(policy, ParserOptions())
}

// Module works like ast.ParseModule but with the Regal parser options applied.
func Module(filename, policy string) (*ast.Module, error) {
	mod, err := ast.ParseModuleWithOpts(filename, policy, ParserOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to parse module: %w", err)
	}

	return mod, nil
}

// PrepareAST prepares the AST to be used as linter input.
func PrepareAST(name string, content string, module *ast.Module) (map[string]any, error) {
	var preparedAST map[string]any

	if err := rio.JSONRoundTrip(module, &preparedAST); err != nil {
		return nil, fmt.Errorf("JSON rountrip failed for module: %w", err)
	}

	preparedAST["regal"] = map[string]any{
		"file": map[string]any{
			"name":  name,
			"lines": strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n"),
		},
	}

	return preparedAST, nil
}
