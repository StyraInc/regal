//nolint:dupl
package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

// Import will return completions for the import keyword when at the start of a line.
type Import struct{}

func (*Import) Run(c *cache.Cache, params types.CompletionParams, _ *Options) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	_, currentLine := completionLineHelper(c, fileURI, params.Position.Line)

	if params.Context.TriggerKind == completion.Invoked && params.Position.Character == 0 {
		return importCompletionItem(params), nil
	}

	// the user must type i before we provide completions
	if params.Position.Line != 0 && strings.HasPrefix(currentLine, "i") {
		return importCompletionItem(params), nil
	}

	return []types.CompletionItem{}, nil
}

func importCompletionItem(params types.CompletionParams) []types.CompletionItem {
	return []types.CompletionItem{
		{
			Label:  "import",
			Kind:   completion.Keyword,
			Detail: "import <path>",
			TextEdit: &types.TextEdit{
				Range: types.Range{
					Start: types.Position{Line: params.Position.Line, Character: 0},
					End:   types.Position{Line: params.Position.Line, Character: params.Position.Character},
				},
				NewText: "import ",
			},
		},
	}
}
