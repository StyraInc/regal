package providers

import (
	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/lsp/types"
)

type Options struct {
	// Builtins is a map of built-in functions to their definitions required in
	// the context of the current completion request.
	Builtins    map[string]*ast.Builtin
	RootURI     string
	Client      types.Client
	RegoVersion ast.RegoVersion
}
