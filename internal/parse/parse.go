package parse

import (
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/ast/json"

	rio "github.com/styrainc/regal/internal/io"
)

// ParserOptions provides the parse options necessary to include location data in AST results.
func ParserOptions() ast.ParserOptions {
	return ast.ParserOptions{
		ProcessAnnotation: true,
		JSONOptions: &json.Options{
			MarshalOptions: json.MarshalOptions{
				IncludeLocation: json.NodeToggle{
					Term:           true,
					Package:        true,
					Comment:        true,
					Import:         true,
					Rule:           true,
					Head:           true,
					Expr:           true,
					SomeDecl:       true,
					Every:          true,
					With:           true,
					Annotations:    true,
					AnnotationsRef: true,
				},
			},
		},
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

// EnhanceAST TODO rename with https://github.com/StyraInc/regal/issues/86.
func EnhanceAST(name string, content string, module *ast.Module) (map[string]any, error) {
	var enhancedAst map[string]any

	if err := rio.JSONRoundTrip(module, &enhancedAst); err != nil {
		return nil, fmt.Errorf("JSON rountrip failed for module: %w", err)
	}

	enhancedAst["regal"] = map[string]any{
		"file": map[string]any{
			"name":  name,
			"lines": strings.Split(content, "\n"),
		},
	}

	return enhancedAst, nil
}
