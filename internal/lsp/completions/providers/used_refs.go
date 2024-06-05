package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

// UsedRefs is a completion provider that provides completions for refs
// that have already been typed into a module.
type UsedRefs struct{}

func (*UsedRefs) Run(c *cache.Cache, params types.CompletionParams, _ *Options) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	lines, currentLine := completionLineHelper(c, fileURI, params.Position.Line)
	if len(lines) < 1 || currentLine == "" {
		return []types.CompletionItem{}, nil
	}

	if !inRuleBody(currentLine) {
		return []types.CompletionItem{}, nil
	}

	words := patternWhiteSpace.Split(strings.TrimSpace(currentLine), -1)
	lastWord := words[len(words)-1]

	refNames, ok := c.GetUsedRefs(fileURI)
	if !ok {
		return []types.CompletionItem{}, nil
	}

	items := []types.CompletionItem{}

	for _, ref := range refNames {
		if !strings.HasPrefix(ref, lastWord) {
			continue
		}

		items = append(items, types.CompletionItem{
			Label:  ref,
			Kind:   completion.Reference,
			Detail: "Existing ref used in module",
			Documentation: &types.MarkupContent{
				Kind:  "markdown",
				Value: `Existing ref used in module`,
			},
			TextEdit: &types.TextEdit{
				Range: types.Range{
					Start: types.Position{
						Line:      params.Position.Line,
						Character: params.Position.Character - uint(len(lastWord)),
					},
					End: types.Position{
						Line: params.Position.Line, Character: params.Position.Character,
					},
				},
				NewText: ref,
			},
		})
	}

	return items, nil
}
