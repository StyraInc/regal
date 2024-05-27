package providers

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
	"github.com/styrainc/regal/internal/lsp/uri"
)

// PackageName will return completions for the package name when starting a new file based on the file's URI.
type PackageName struct{}

func (*PackageName) Run(c *cache.Cache, params types.CompletionParams, opts *Options) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	lines, currentLine := completionLineHelper(c, fileURI, params.Position.Line)
	if len(lines) < 1 || currentLine == "" {
		return []types.CompletionItem{}, nil
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
	suggestedPackageName := filepath.Base(filepath.Dir(path))

	if opts != nil {
		if trimmed := strings.TrimPrefix(fileURI, opts.RootURI); trimmed != fileURI {
			dir := filepath.Dir(trimmed)
			noLeadingSlash := strings.TrimPrefix(dir, "/")
			noPeriods := strings.ReplaceAll(noLeadingSlash, ".", "_")
			suggestedPackageName = strings.ReplaceAll(noPeriods, "/", ".")
		}
	}

	if !strings.HasPrefix(currentLine, "p") {
		return nil, nil
	}

	return []types.CompletionItem{
		{
			Label:  "package " + suggestedPackageName,
			Detail: "suggested package name based on directory",
			Kind:   completion.Folder,
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
				NewText: fmt.Sprintf("package %s\n\n", suggestedPackageName),
			},
		},
	}, nil
}
