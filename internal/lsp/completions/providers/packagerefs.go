package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/symbols"
)

// PackageRefs is a completion provider that returns completions when importing packages.
type PackageRefs struct{}

func (*PackageRefs) Run(c *cache.Cache, params types.CompletionParams, _ *Options) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	lines, currentLine := completionLineHelper(c, fileURI, params.Position.Line)
	if len(lines) < 1 || currentLine == "" {
		return []types.CompletionItem{}, nil
	}

	// this completion provider applies on lines with import at the start
	if !strings.HasPrefix(currentLine, "import ") {
		return nil, nil
	}

	thisFileReferences := c.GetFileRefs(fileURI)
	otherFilePackages := make(map[string]types.Ref)

	for file, refs := range c.GetAllFileRefs() {
		if file == fileURI {
			continue
		}

		for key, ref := range refs {
			if ref.Kind != types.Package {
				continue
			}

			// don't suggest packages that are defined in this file
			if _, ok := thisFileReferences[key]; ok {
				continue
			}

			otherFilePackages[key] = ref
		}
	}

	words := strings.Split(currentLine, " ")
	lastWord := words[len(words)-1]

	items := make([]types.CompletionItem, 0)

	for _, item := range otherFilePackages {
		if !strings.HasPrefix(item.Label, lastWord) {
			continue
		}

		items = append(items, types.CompletionItem{
			Label:  item.Label,
			Kind:   uint(symbols.Module), // for now, only modules are returned
			Detail: "Rego package",
			Documentation: &types.MarkupContent{
				Kind:  "markdown",
				Value: item.Description,
			},
			TextEdit: &types.TextEdit{
				Range: types.Range{
					Start: types.Position{
						Line:      params.Position.Line,
						Character: 7,
					},
					End: types.Position{
						Line:      params.Position.Line,
						Character: uint(len(currentLine)),
					},
				},
				NewText: item.Label,
			},
		})
	}

	return items, nil
}
