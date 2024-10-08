package providers

import (
	"context"
	"errors"
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/hover"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

type BuiltIns struct{}

func (*BuiltIns) Name() string {
	return "builtins"
}

func (*BuiltIns) Run(
	_ context.Context,
	c *cache.Cache,
	params types.CompletionParams,
	opts *Options,
) ([]types.CompletionItem, error) {
	if opts == nil {
		return nil, errors.New("builtins provider requires options")
	}

	fileURI := params.TextDocument.URI

	lines, currentLine := completionLineHelper(c, fileURI, params.Position.Line)

	if len(lines) < 1 || currentLine == "" {
		return []types.CompletionItem{}, nil
	}

	if !inRuleBody(currentLine) {
		return []types.CompletionItem{}, nil
	}

	// default rules cannot contain calls
	if strings.HasPrefix(strings.TrimSpace(currentLine), "default ") {
		return []types.CompletionItem{}, nil
	}

	words := patternWhiteSpace.Split(strings.TrimSpace(currentLine), -1)
	lastWord := words[len(words)-1]

	items := []types.CompletionItem{}

	for _, builtIn := range opts.Builtins {
		key := builtIn.Name

		if builtIn.Infix != "" {
			continue
		}

		if builtIn.IsDeprecated() {
			continue
		}

		if !strings.HasPrefix(key, lastWord) {
			continue
		}

		items = append(items, types.CompletionItem{
			Label:  key,
			Kind:   completion.Function,
			Detail: "built-in function",
			Documentation: &types.MarkupContent{
				Kind:  "markdown",
				Value: hover.CreateHoverContent(builtIn),
			},
			TextEdit: &types.TextEdit{
				Range: types.Range{
					Start: types.Position{
						Line:      params.Position.Line,
						Character: params.Position.Character - uint(len(lastWord)),
					},
					End: types.Position{
						Line:      params.Position.Line,
						Character: params.Position.Character,
					},
				},
				NewText: key,
			},
		})
	}

	return items, nil
}
