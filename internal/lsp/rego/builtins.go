package rego

import (
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

// BuiltinsForCapabilities returns a list of builtins from the provided capabilities.
func BuiltinsForCapabilities(capabilities *ast.Capabilities) map[string]*ast.Builtin {
	m := make(map[string]*ast.Builtin)
	for _, b := range capabilities.Builtins {
		m[b.Name] = b
	}

	return m
}

func BuiltinCategory(builtin *ast.Builtin) (category string) {
	if len(builtin.Categories) == 0 {
		if s := strings.Split(builtin.Name, "."); len(s) > 1 {
			category = s[0]
		} else {
			category = builtin.Name
		}
	} else {
		category = builtin.Categories[0]
	}

	return category
}
