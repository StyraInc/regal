package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/hover"
	"github.com/styrainc/regal/internal/lsp/rego"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

type BuiltIns struct{}

func (*BuiltIns) Run(c *cache.Cache, params types.CompletionParams, _ *Options) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	lines, currentLine := completionLineHelper(c, fileURI, params.Position.Line)
	if len(lines) < 1 || currentLine == "" {
		return []types.CompletionItem{}, nil
	}

	// TODO: Share and improve this logic, currently shared with the rulerefs provider
	if !strings.Contains(currentLine, " if ") && // if after if keyword
		!strings.Contains(currentLine, " contains ") && // if after contains
		!strings.Contains(currentLine, " else ") && // if after else
		!strings.Contains(currentLine, "= ") && // if after assignment
		!patternRuleBody.MatchString(currentLine) { // if in rule body
		return nil, nil
	}

	words := patternWhiteSpace.Split(strings.TrimSpace(currentLine), -1)
	lastWord := words[len(words)-1]

	items := []types.CompletionItem{}

	for key, builtIn := range rego.BuiltIns {
		if builtIn.Infix != "" {
			continue
		}

		if builtIn.IsDeprecated() {
			continue
		}

		if strings.HasPrefix(key, lastWord) {
			items = append(items, types.CompletionItem{
				Label:  key,
				Kind:   completion.Function,
				Detail: "",
				Documentation: &types.MarkupContent{
					Kind:  "markdown",
					Value: hover.CreateHoverContent(builtIn),
				},
			})
		}
	}

	return items, nil
}
