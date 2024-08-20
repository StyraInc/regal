package providers

import (
	"github.com/open-policy-agent/opa/ast"
	"github.com/styrainc/regal/internal/lsp/clients"
)

type Options struct {
	ClientIdentifier clients.Identifier
	RootURI          string
	Capabilities     *ast.Capabilities
}
