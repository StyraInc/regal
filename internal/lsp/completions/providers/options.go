package providers

import (
	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/lsp/clients"
)

type Options struct {
	ClientIdentifier clients.Identifier
	RootURI          string
	// Builtins is a map of built-in functions to their definitions required in
	// the context of the current completion request.
	Builtins map[string]*ast.Builtin
}
