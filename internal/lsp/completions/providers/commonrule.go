//nolint:dupl
package providers

import (
	"fmt"
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

// CommonRule will return completions for new rules based on common Rego rule names.
type CommonRule struct{}

func (*CommonRule) Run(c *cache.Cache, params types.CompletionParams, _ *Options) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	_, currentLine := completionLineHelper(c, fileURI, params.Position.Line)

	if patternRuleBody.MatchString(currentLine) { // if in rule body
		return []types.CompletionItem{}, nil
	}

	words := patternWhiteSpace.Split(currentLine, -1)
	if len(words) != 1 {
		return []types.CompletionItem{}, nil
	}

	// if the file already contains a rule with the same name, we do not want to
	// suggest it again. In order to be able to do this later, we need to record
	// all the existing rules in the file.
	existingRules := make(map[string]struct{})

	for _, ref := range c.GetFileRefs(fileURI) {
		if ref.Kind == types.Rule || ref.Kind == types.ConstantRule || ref.Kind == types.Function {
			parts := strings.Split(ref.Label, ".")
			existingRules[parts[len(parts)-1]] = struct{}{}
		}
	}

	lastWord := strings.TrimSpace(currentLine)

	var label string

	var newText string

	for _, word := range []string{"allow", "deny"} {
		if strings.HasPrefix(word, lastWord) {
			// if the rule is defined, we can skip it as it'll be suggested by
			// another provider
			if _, ok := existingRules[word]; ok {
				return []types.CompletionItem{}, nil
			}

			label = word
			newText = word + " "

			break
		}
	}

	if label == "" {
		return []types.CompletionItem{}, nil
	}

	return []types.CompletionItem{
		{
			Label: label,
			Kind:  completion.Snippet,
			Documentation: &types.MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("%q is a common rule name", label),
			},
			TextEdit: &types.TextEdit{
				Range: types.Range{
					Start: types.Position{
						Line:      params.Position.Line,
						Character: params.Position.Character - uint(len(lastWord)),
					},
					End: types.Position{
						Line:      params.Position.Line,
						Character: uint(len(currentLine)),
					},
				},
				NewText: newText,
			},
		},
	}, nil
}
