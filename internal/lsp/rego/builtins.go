package rego

import (
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
)

// BuiltinsForCapabilities returns a list of builtins from the provided capabilities.
func BuiltinsForCapabilities(capabilities *ast.Capabilities) map[string]*ast.Builtin {
	m := make(map[string]*ast.Builtin, len(capabilities.Builtins))
	for _, b := range capabilities.Builtins {
		m[b.Name] = b
	}

	return m
}

func BuiltinCategory(builtin *ast.Builtin) (category string) {
	if len(builtin.Categories) == 0 {
		category = builtin.Name
		if i := strings.Index(builtin.Name, "."); i > -1 {
			category = builtin.Name[:i]
		}
	} else {
		category = builtin.Categories[0]
	}

	return category
}
