package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

// RuleFromImportedPackageRefs is a completion provider that returns completions for
// rules found in already imported packages.
type RuleFromImportedPackageRefs struct{}

func (*RuleFromImportedPackageRefs) Run(
	c *cache.Cache,
	params types.CompletionParams,
	_ *Options,
) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	lines, currentLine := completionLineHelper(c, fileURI, params.Position.Line)
	if len(lines) < 1 || currentLine == "" {
		return []types.CompletionItem{}, nil
	}

	// TODO: Share and improve this logic, currently shared with the builtins provider
	if !strings.Contains(currentLine, " if ") && // if after if keyword
		!strings.Contains(currentLine, " contains ") && // if after contains
		!strings.Contains(currentLine, " else ") && // if after else
		!strings.Contains(currentLine, "= ") && // if after assignment
		!strings.HasPrefix(currentLine, "  ") { // if in rule body
		return nil, nil
	}

	// some version of a parsed mod is needed here to filter refs to suggest
	// based on import statements
	mod, ok := c.GetModule(fileURI)
	if !ok {
		return nil, nil
	}

	refsFromImports := make(map[string]types.Ref)

	for file, refs := range c.GetAllFileRefs() {
		if file == fileURI {
			continue
		}

		for key, ref := range refs {
			// we are not interested in packages here, only the rules
			if ref.Kind == types.Package {
				continue
			}

			// don't suggest refs of "private" rules, even
			// if only just by naming convention
			if strings.Contains(ref.Label, "._") {
				continue
			}

			isFromImportedPackage := false

			for _, i := range mod.Imports {
				if strings.HasPrefix(key, i.Path.String()) {
					isFromImportedPackage = true

					break
				}
			}

			if !isFromImportedPackage {
				continue
			}

			refsFromImports[key] = ref
		}
	}

	items := make([]types.CompletionItem, 0)

	for _, ref := range refsFromImports {
		symbol := completion.Variable
		detail := "Rule"

		switch {
		case ref.Kind == types.ConstantRule:
			symbol = completion.Constant
			detail = "Constant Rule"
		case ref.Kind == types.Function:
			symbol = completion.Function
			detail = "Function"
		}

		packageAndRule := labelToPackageAndRule(ref.Label)

		startChar := params.Position.Character -
			uint(len(strings.Split(currentLine, " ")[len(strings.Split(currentLine, " "))-1]))

		items = append(items, types.CompletionItem{
			Label:  packageAndRule,
			Kind:   symbol,
			Detail: detail,
			Documentation: &types.MarkupContent{
				Kind:  "markdown",
				Value: ref.Description,
			},
			TextEdit: &types.TextEdit{
				Range: types.Range{
					Start: types.Position{
						Line:      params.Position.Line,
						Character: startChar,
					},
					End: types.Position{
						Line:      params.Position.Line,
						Character: uint(len(currentLine)),
					},
				},
				NewText: packageAndRule,
			},
		})
	}

	return items, nil
}

func labelToPackageAndRule(label string) string {
	parts := strings.Split(label, ".")
	partCount := len(parts)

	// a ref should be at least three parts, data.package.rule_from_package
	// if it is not, then we can't provide a valid new text and return the
	// full label as a fallback
	if partCount < 3 {
		return label
	}

	// take the last two parts of the ref, package and rule
	return parts[partCount-2] + "." + parts[partCount-1]
}
