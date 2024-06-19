package providers

import "github.com/styrainc/regal/internal/lsp/clients"

type Options struct {
	ClientIdentifier clients.Identifier
	RootURI          string
}
