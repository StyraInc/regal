package providers

import (
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/pkg/config"
)

type Options struct {
	RootURI          string
	ClientIdentifier clients.Identifier
	Ignore           config.Ignore
}
