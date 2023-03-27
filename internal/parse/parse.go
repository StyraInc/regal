package parse

import "github.com/open-policy-agent/opa/ast"

// ParserOptions provides the parse options necessary to include location data in AST results.
func ParserOptions() ast.ParserOptions {
	return ast.ParserOptions{
		ProcessAnnotation: true,
		JSONOptions: &ast.JSONOptions{
			MarshalOptions: ast.JSONMarshalOptions{
				IncludeLocation: ast.NodeToggle{
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
