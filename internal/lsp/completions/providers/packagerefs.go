package providers

import (
	"slices"
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
	otherFilePackages := make(map[string]types.Ref)

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

			otherFilePackages[key] = ref
		}
	}

	// partialRefs is a generated list of package names generated from
	// longer names. For example, if the packages data.foo.bar and data.foo.baz
	// are defined, an author should still be able to import data.foo.
	partialRefs := make(map[string]types.Ref)

	for key := range otherFilePackages {
		parts := strings.Split(key, ".")
		// starting at 1 to skip 'data'
		for i := 1; i < len(parts)-1; i++ {
			partialKey := strings.Join(parts[:i+1], ".")
			// only insert the new partial key if there is no full package
			// ref that matches it
			if _, ok := partialRefs[partialKey]; !ok {
				partialRefs[partialKey] = types.Ref{
					Label:       partialKey,
					Description: "See sub packages for more information",
				}
			}
		}
	}

	// refs are grouped by 'depth', where depth is the number of dots in the
	// ref string. This is a simplification, but allows shorted, higher level
	// refs to be suggested first.
	byDepth := make(map[int]map[string]types.Ref)

	for _, item := range otherFilePackages {
		depth := strings.Count(item.Label, ".")

		if _, ok := byDepth[depth]; !ok {
			byDepth[depth] = make(map[string]types.Ref)
		}

		byDepth[depth][item.Label] = item
	}

	// add partial refs to the byDepth map in case they are not defined
	// as full refs in files.
	for _, item := range partialRefs {
		depth := strings.Count(item.Label, ".")

		if _, ok := byDepth[depth]; !ok {
			byDepth[depth] = make(map[string]types.Ref)
		}

		// only add partial refs where no top level ref exists.
		if _, ok := byDepth[depth][item.Label]; ok {
			continue
		}

		byDepth[depth][item.Label] = item
	}

	// items will be shown in order from shallowest to deepest
	depths := make([]int, 0)
	for k := range byDepth {
		depths = append(depths, k)
	}

	slices.Sort(depths)

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
