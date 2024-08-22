package rego

import (
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/ast"
)

var builtInsLock = &sync.RWMutex{}                          // nolint:gochecknoglobals
var builtIns = builtinMap(ast.CapabilitiesForThisVersion()) //nolint:gochecknoglobals

// Update updates the builtins database with the provided capabilities.
func UpdateBuiltins(caps *ast.Capabilities) {
	builtInsLock.Lock()
	builtIns = builtinMap(caps)
	builtInsLock.Unlock()
}

func GetBuiltins() map[string]*ast.Builtin {
	builtInsLock.RLock()
	defer builtInsLock.RUnlock()
	return builtIns
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
