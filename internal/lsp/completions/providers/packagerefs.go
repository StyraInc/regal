package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
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

	words := strings.Split(currentLine, " ")
	lastWord := words[len(words)-1]

	thisFileReferences := c.GetFileRefs(fileURI)
	refsForContext := make(map[string]types.Ref)

	// filter out the packages that have the last word as a prefix.
	for file, refs := range c.GetAllFileRefs() {
		if file == fileURI {
			continue
		}

		for key, ref := range refs {
			// this provider is only used to suggest items where a package is suitable.
			if ref.Kind != types.Package {
				continue
			}

			// don't suggest packages that are defined in this file
			if _, ok := thisFileReferences[key]; ok {
				continue
			}

			// only suggest packages that match the last word
			// the user has typed.
			if !strings.HasPrefix(ref.Label, lastWord) {
				continue
			}

			refsForContext[key] = ref
		}
	}

	// refsForContext is now supplemented with a generated list of package names
	// from longer packages. For example, if the packages data.foo.bar and data.foo.baz
	// are defined, an author should still be able to import data.foo.
	for key := range refsForContext {
		parts := strings.Split(key, ".")
		// starting at 1 to skip 'data'
		for i := 1; i < len(parts)-1; i++ {
			partialKey := strings.Join(parts[:i+1], ".")
			// only insert the new partial key if there is no full package
			// ref that matches it
			if _, ok := refsForContext[partialKey]; !ok {
				refsForContext[partialKey] = types.Ref{
					Label:       partialKey,
					Description: "See sub packages for more information",
				}
			}
		}
	}

	// refs are grouped by 'depth', where depth is the number of dots in the
	// ref string. This is a simplification, but allows shorted, higher level
	// refs to be suggested first.
	depths, byDepth := groupKeyedRefsByDepth(refsForContext)

	items := make([]types.CompletionItem, 0)
	for _, depth := range depths {
		// items are added in groups of depth until there more then 10 items.
		if len(items) > 10 {
			continue
		}

		itemsForDepth, ok := byDepth[depth]
		if !ok {
			continue
		}

		for _, item := range itemsForDepth {
			items = append(items, types.CompletionItem{
				Label:  item.Label,
				Kind:   completion.Module, // for now, only modules are returned
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
	}

	return items, nil
}
