package completions

import (
	"fmt"

	"github.com/open-policy-agent/opa/storage"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/completions/providers"
	"github.com/styrainc/regal/internal/lsp/rego"
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
	Name() string
}

func NewManager(c *cache.Cache, opts *ManagerOptions) *Manager {
	return &Manager{c: c, opts: opts}
}

func NewDefaultManager(c *cache.Cache, store storage.Store) *Manager {
	m := NewManager(c, &ManagerOptions{})

	m.RegisterProvider(&providers.BuiltIns{})
	m.RegisterProvider(&providers.PackageRefs{})
	m.RegisterProvider(&providers.RuleHead{})
	m.RegisterProvider(&providers.RuleHeadKeyword{})
	m.RegisterProvider(&providers.Input{})
	m.RegisterProvider(&providers.UsedRefs{})

	m.RegisterProvider(providers.NewPolicy(store))

	return m
}

func (m *Manager) Run(params types.CompletionParams, opts *providers.Options) ([]types.CompletionItem, error) {
	completions := make(map[string][]types.CompletionItem)

	if m.isInsideOfComment(params) {
		// Exit early if caret position is inside a comment. We currently don't have any provider
		// where doing completions inside of a comment makes much sense. Behavior is also editor-specific:
		// - Zed: always on, with no way to disable
		// - VSCode: disabled but can be enabled with "editor.quickSuggestions.comments" setting
		return []types.CompletionItem{}, nil
	}

	for _, provider := range m.providers {
		providerCompletions, err := provider.Run(m.c, params, opts)
		if err != nil {
			return nil, fmt.Errorf("error running completion provider: %w", err)
		}

		for _, completion := range providerCompletions {
			// If a provider returns a mandatory completion, return it immediately
			// as it is the only completion that should be shown.
			if completion.Mandatory {
				return []types.CompletionItem{completion}, nil
			}

			completions[completion.Label] = append(completions[completion.Label], completion)
		}
	}

	var completionsList []types.CompletionItem

	for _, completionItems := range completions {
		if len(completionItems) < 2 {
			completionsList = append(completionsList, completionItems...)

			continue
		}

		maxRank := 0

		for _, completion := range completionItems {
			if completion.Regal == nil {
				continue
			}

			if rank := rankProvider(completion.Regal.Provider); rank > maxRank {
				maxRank = rank
			}
		}

		for _, completion := range completionItems {
			if completion.Regal == nil {
				completionsList = append(completionsList, completion)

				continue
			}

			if rank := rankProvider(completion.Regal.Provider); rank == maxRank {
				completionsList = append(completionsList, completion)
			}
		}
	}

	for i := range completionsList {
		completionsList[i].Regal = nil
	}

	return completionsList, nil
}

func rankProvider(provider string) int {
	switch provider {
	case "rulerefs":
		return 100
	case "usedrefs":
		return 90
	default:
		return 0
	}
}

func (m *Manager) RegisterProvider(provider Provider) {
	m.providers = append(m.providers, provider)
}

func (m *Manager) isInsideOfComment(params types.CompletionParams) bool {
	if module, ok := m.c.GetModule(params.TextDocument.URI); ok {
		for _, comment := range module.Comments {
			cp := rego.PositionFromLocation(comment.Location)

			if cp.Line == params.Position.Line {
				if cp.Character <= params.Position.Character {
					return true
				}
			}
		}
	}

	return false
}
