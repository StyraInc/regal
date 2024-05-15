package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
)

// Package will return completions for the package keyword when starting a new file.
type Package struct{}

func (*Package) Run(c *cache.Cache, params types.CompletionParams) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	fileContents, ok := c.GetFileContents(fileURI)
	if !ok {
		// if the file contents is missing then we can't provide completions
		return []types.CompletionItem{}, nil
	}

	lines := strings.Split(fileContents, "\n")
	if len(lines) < 1 {
		return []types.CompletionItem{}, nil
	}

	for i, line := range lines {
		if i < int(params.Position.Line) && strings.HasPrefix(line, "package ") {
			// if there is already a package statement in the file then we don't provide any more completions
			return []types.CompletionItem{}, nil
		}
	}

	// if we can't confirm that the user has package statement on the line then we don't provide completions
	if len(lines)+1 < int(params.Position.Line) {
		return []types.CompletionItem{}, nil
	}

	// if not on the first line, the user must type p before we provide completions
	if params.Position.Line != 0 && !strings.HasPrefix(lines[params.Position.Line], "p") {
		return []types.CompletionItem{}, nil
	}

	return []types.CompletionItem{
		{
			Label:  "package",
			Kind:   14, // 14 is the kind for keyword
			Detail: "package <package-name>",
			TextEdit: &types.TextEdit{
				Range: types.Range{
					Start: types.Position{
						Line:      params.Position.Line,
						Character: 0,
					},
					End: types.Position{
						Line:      params.Position.Line,
						Character: params.Position.Character,
					},
				},
				NewText: "package ",
			},
		},
	}, nil
}
