package rego

import (
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/ast"
)

var BuiltInsLock = &sync.RWMutex{}                          // nolint:gochecknoglobals
var BuiltIns = builtinMap(ast.CapabilitiesForThisVersion()) //nolint:gochecknoglobals

func UpdateBuiltins(caps *ast.Capabilities) {
	BuiltInsLock.Lock()
	defer BuiltInsLock.Unlock()
	BuiltIns = builtinMap(caps)
}

func builtinMap(caps *ast.Capabilities) map[string]*ast.Builtin {
	m := make(map[string]*ast.Builtin)
	for _, b := range caps.Builtins {
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
