//nolint:dupl
package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

// Default will return completions for the default keyword when at the start of a line.
type Default struct{}

func (*Default) Run(c *cache.Cache, params types.CompletionParams, _ *Options) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	_, currentLine := completionLineHelper(c, fileURI, params.Position.Line)

	if params.Context.TriggerKind == completion.Invoked && params.Position.Character == 0 {
		return defaultCompletionItem(params), nil
	}

	// the user must type d before we provide completions
	if params.Position.Line != 0 && strings.HasPrefix(currentLine, "d") {
		return defaultCompletionItem(params), nil
	}

	return []types.CompletionItem{}, nil
}

func defaultCompletionItem(params types.CompletionParams) []types.CompletionItem {
	return []types.CompletionItem{
		{
			Label:  "default",
			Kind:   completion.Keyword,
			Detail: "default <rule-name> := <value>",
			TextEdit: &types.TextEdit{
				Range: types.Range{
					Start: types.Position{Line: params.Position.Line, Character: 0},
					End:   types.Position{Line: params.Position.Line, Character: params.Position.Character},
				},
				NewText: "default ",
			},
		},
	}
}
