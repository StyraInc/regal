package providers

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/uri"
)

// PackageName will return completions for the package name when starting a new file based on the file's URI.
type PackageName struct{}

func (p *PackageName) Run(c *cache.Cache, params types.CompletionParams) ([]types.CompletionItem, error) {

	fileURI := params.TextDocument.URI
	fileContents, ok := c.GetFileContents(fileURI)
	if !ok {
		// if the file contents is missing then we can't provide completions
		return nil, nil
	}

	lines := strings.Split(fileContents, "\n")
	if len(lines) < 1 {
		return nil, nil
	}

	for i, line := range lines {
		if i < int(params.Position.Line) && strings.HasPrefix(line, "package ") {
			// if there is already a package statement in the file then we don't provide any more completions
			return nil, nil
		}
	}

	// if we can't confirm that the user has package statement on the line then we don't provide completions
	if len(lines)+1 < int(params.Position.Line) {
		return nil, nil
	}

	path := uri.ToPath(clients.IdentifierGeneric, fileURI)
	dir := filepath.Base(filepath.Dir(path))

	if !strings.HasPrefix(lines[params.Position.Line], "p") {
		return nil, nil
	}

	return []types.CompletionItem{
		{
			Label:  fmt.Sprintf("package %s", dir),
			Detail: "suggested package name based on directory",
			Kind:   19, // 19 is the kind for a folder
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
				NewText: fmt.Sprintf("package %s\n\n", dir),
			},
		},
	}, nil
}
