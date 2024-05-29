package completions

import (
	"fmt"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/completions/providers"
	"github.com/styrainc/regal/internal/lsp/types"
)

type Manager struct {
	c         *cache.Cache
	opts      *ManagerOptions
	providers []Provider
}

type ManagerOptions struct{}

type Provider interface {
	Run(*cache.Cache, types.CompletionParams, *providers.Options) ([]types.CompletionItem, error)
}

func NewManager(c *cache.Cache, opts *ManagerOptions) *Manager {
	return &Manager{c: c, opts: opts}
}

func NewDefaultManager(c *cache.Cache) *Manager {
	m := NewManager(c, &ManagerOptions{})

	m.RegisterProvider(&providers.Package{})
	m.RegisterProvider(&providers.PackageName{})
	m.RegisterProvider(&providers.Default{})
	m.RegisterProvider(&providers.Import{})
	m.RegisterProvider(&providers.BuiltIns{})
	m.RegisterProvider(&providers.RegoV1{})
	m.RegisterProvider(&providers.PackageRefs{})
	m.RegisterProvider(&providers.RuleRefs{})
	m.RegisterProvider(&providers.RuleHead{})

	return m
}

func (m *Manager) Run(params types.CompletionParams, opts *providers.Options) ([]types.CompletionItem, error) {
	var completions []types.CompletionItem

	for _, provider := range m.providers {
		providerCompletions, err := provider.Run(m.c, params, opts)
		if err != nil {
			return nil, fmt.Errorf("error running completion provider: %w", err)
		}

		if len(providerCompletions) > 0 {
			completions = append(completions, providerCompletions...)
		}
	}

	return completions, nil
}

func (m *Manager) RegisterProvider(provider Provider) {
	m.providers = append(m.providers, provider)
}
