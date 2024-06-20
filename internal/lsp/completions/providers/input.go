package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

// Input is a simple completion provider that returns the input keyword as a completion item
// at suitable times.
type Input struct{}

func (*Input) Name() string {
	return "input"
}

func (*Input) Run(c *cache.Cache, params types.CompletionParams, _ *Options) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	lines, currentLine := completionLineHelper(c, fileURI, params.Position.Line)
	if len(lines) < 1 || currentLine == "" {
		return []types.CompletionItem{}, nil
	}

	if !inRuleBody(currentLine) {
		return []types.CompletionItem{}, nil
	}

	words := patternWhiteSpace.Split(strings.TrimSpace(currentLine), -1)
	lastWord := words[len(words)-1]

	items := []types.CompletionItem{}

	//nolint:gocritic
	if strings.HasPrefix("input", lastWord) {
		items = append(items, types.CompletionItem{
			Label:  "input",
			Kind:   completion.Keyword,
			Detail: "document",
			Documentation: &types.MarkupContent{
				Kind: "markdown",
				Value: `# input

'input' refers to the input document being evaluated.
It is a special keyword that allows you to access the data sent to OPA at evaluation time.

To see more examples of how to use 'input', check out the
[policy language documentation](https://www.openpolicyagent.org/docs/latest/policy-language/).

You can also experiment with input in the [Rego Playground](https://play.openpolicyagent.org/).
`,
			},
			TextEdit: &types.TextEdit{
				Range: types.Range{
					Start: types.Position{
						Line:      params.Position.Line,
						Character: params.Position.Character - uint(len(lastWord)),
					},
					End: params.Position,
				},
				NewText: "input",
			},
		})
	}

	return items, nil
}
