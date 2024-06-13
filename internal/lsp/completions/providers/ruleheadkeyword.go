//nolint:dupl
package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

// RuleHeadKeyword will return completions for the keywords when starting a new rule.
// The current cases are supported:
// - [rule-name] if
// - [rule-name] contains
// - [rule-name] contains if
// These completions are mandatory, that means they are the only ones to be shown.
type RuleHeadKeyword struct{}

func (*RuleHeadKeyword) Run(c *cache.Cache, params types.CompletionParams, _ *Options) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	_, currentLine := completionLineHelper(c, fileURI, params.Position.Line)

	if patternRuleBody.MatchString(currentLine) { // if in rule body
		return []types.CompletionItem{}, nil
	}

	words := patternWhiteSpace.Split(currentLine, -1)
	if len(words) < 2 {
		return []types.CompletionItem{}, nil
	}

	firstWord := strings.TrimSpace(words[0])
	if firstWord == "package" || firstWord == "import" {
		return []types.CompletionItem{}, nil
	}

	lastWord := words[len(words)-1]

	const keyWdContains = "contains"

	const keyWdIf = "if"

	var label string

	switch {
	// suggest contains after the name of the rule in the rule head
	//nolint:gocritic
	case len(words) == 2 && strings.HasPrefix(keyWdContains, lastWord):
		label = "contains"
	// suggest if at the end of the rule head
	case len(words) == 4 && words[1] == keyWdContains:
		label = keyWdIf
	// suggest if after the rule name
	//nolint:gocritic
	case len(words) == 2 && strings.HasPrefix(keyWdIf, lastWord):
		label = keyWdIf
	}

	if label == "" {
		return []types.CompletionItem{}, nil
	}

	return []types.CompletionItem{
		{
			Label: label,
			Kind:  completion.Keyword,
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
				NewText: label + " ",
			},
			Mandatory: true,
		},
	}, nil
}
