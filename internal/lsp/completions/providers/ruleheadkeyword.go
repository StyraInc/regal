//nolint:dupl
package providers

import (
	"context"
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

// RuleHeadKeyword will return completions for the keywords when starting a new rule.
// The current cases are supported:
// - [rule-name] if
// - [rule-name] :=
// - [rule-name] contains
// - [rule-name] contains if
// These completions are mandatory, that means they are the only ones to be shown.
type RuleHeadKeyword struct{}

func (*RuleHeadKeyword) Name() string {
	return "ruleheadkeyword"
}

func (*RuleHeadKeyword) Run(
	_ context.Context,
	c *cache.Cache,
	params types.CompletionParams,
	_ *Options,
) ([]types.CompletionItem, error) {
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

	const keyWdAssign = ":="

	mandatory := false
	keywords := map[string]bool{
		keyWdIf:       true,
		keyWdContains: true,
		keyWdAssign:   true,
	}

	if lastWord != "" {
		mandatory = true
		keywords = map[string]bool{
			keyWdIf:       false,
			keyWdContains: false,
		}

		switch {
		// suggest the assignment operator after the name of the rule
		case len(words) == 2 && strings.HasPrefix(keyWdAssign, lastWord):
			keywords[keyWdAssign] = true
		// suggest contains after the name of the rule in the rule head
		//nolint:gocritic
		case len(words) == 2 && strings.HasPrefix(keyWdContains, lastWord):
			keywords[keyWdContains] = true
		// suggest if at the end of the rule head
		case len(words) == 4 && words[1] == keyWdContains:
			keywords[keyWdIf] = true
		// suggest if after the rule name
		//nolint:gocritic
		case len(words) == 2 && strings.HasPrefix(keyWdIf, lastWord):
			keywords[keyWdIf] = true
		}
	}

	ret := make([]types.CompletionItem, 0)

	for k, v := range keywords {
		if !v {
			continue
		}

		ret = append(ret, types.CompletionItem{
			Label: k,
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
				NewText: k + " ",
			},
			Mandatory: mandatory,
		})
	}

	return ret, nil
}
