package providers

import (
	"context"
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

// RuleHead is a completion provider that returns completions for
// rules found in the same package at the start of a line, so
// when adding new heads, the user can easily add new ones.
type RuleHead struct{}

func (*RuleHead) Name() string {
	return "rulehead"
}

func (*RuleHead) Run(
	_ context.Context,
	c *cache.Cache,
	params types.CompletionParams,
	_ *Options,
) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	lines, currentLine := completionLineHelper(c, fileURI, params.Position.Line)
	if len(lines) < 1 {
		return []types.CompletionItem{}, nil
	}

	if currentLine != "" {
		words := patternWhiteSpace.Split(currentLine, -1)

		// this provider only suggests rules at the start of a line while
		// typing the first word
		if len(words) != 1 {
			return []types.CompletionItem{}, nil
		}

		// if the first word is not at the start of the line, then we
		// assume the cursor is in a rule body and we exit
		if strings.Index(currentLine, words[0]) != 0 {
			return []types.CompletionItem{}, nil
		}
	}

	// some version of a parsed mod is needed here to filter refs to suggest
	// based on import statements
	mod, ok := c.GetModule(fileURI)
	if !ok {
		return nil, nil
	}

	modPrefix := mod.Package.Path.String() + "."

	refsFromPackage := make(map[string]types.Ref)

	// we gather refs from other files in case the package has been defined
	// in more than one file
	for _, refs := range c.GetAllFileRefs() {
		for key, ref := range refs {
			// this provider only suggests rules
			if ref.Kind != types.Rule && ref.Kind != types.ConstantRule && ref.Kind != types.Function {
				continue
			}

			// only rules from the current package are suggested
			if !strings.HasPrefix(key, modPrefix) {
				continue
			}

			refsFromPackage[strings.TrimPrefix(key, modPrefix)] = ref
		}
	}

	items := make([]types.CompletionItem, 0)

	for key, ref := range refsFromPackage {
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

		items = append(items, types.CompletionItem{
			Label:  key,
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
						Character: 0,
					},
					End: types.Position{
						Line:      params.Position.Line,
						Character: uint(len(currentLine)),
					},
				},
				NewText: key,
			},
		})
	}

	return items, nil
}
