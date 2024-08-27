package providers

import (
	"fmt"
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/hover"
	"github.com/styrainc/regal/internal/lsp/rego"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

type BuiltIns struct{}

func (*BuiltIns) Name() string {
	return "builtins"
}

func (*BuiltIns) Run(c *cache.Cache, params types.CompletionParams, opt *Options) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	fmt.Printf("CCC22 %+v %+v\n", params, opt)

	lines, currentLine := completionLineHelper(c, fileURI, params.Position.Line)

	for i, l := range lines {
		fmt.Printf("CCC28 line %d: %s\n", i, l)
	}
	fmt.Printf("CCC29: current line: %s\n", currentLine)

	if len(lines) < 1 || currentLine == "" {
		fmt.Printf("CCC27\n")
		return []types.CompletionItem{}, nil
	}

	if !inRuleBody(currentLine) {
		fmt.Printf("CCC32\n")
		return []types.CompletionItem{}, nil
	}

	fmt.Printf("CCC34 currentLine %+v\n", currentLine)

	words := patternWhiteSpace.Split(strings.TrimSpace(currentLine), -1)
	lastWord := words[len(words)-1]

	fmt.Printf("CCC47 lastWord: %s\n", lastWord)

	items := []types.CompletionItem{}

	bis := rego.GetBuiltins()

	for _, builtIn := range bis {
		key := builtIn.Name

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
				Detail: "built-in function",
				Documentation: &types.MarkupContent{
					Kind:  "markdown",
					Value: hover.CreateHoverContent(builtIn),
				},
				TextEdit: &types.TextEdit{
					Range: types.Range{
						Start: types.Position{
							Line:      params.Position.Line,
							Character: params.Position.Character - uint(len(lastWord)),
						},
						End: types.Position{
							Line:      params.Position.Line,
							Character: params.Position.Character,
						},
					},
					NewText: key,
				},
			})
		}
	}

	return items, nil
}
