package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/hover"
	"github.com/styrainc/regal/internal/lsp/rego"
	"github.com/styrainc/regal/internal/lsp/types"
)

type BuiltIns struct{}

func (*BuiltIns) Run(c *cache.Cache, params types.CompletionParams) ([]types.CompletionItem, error) {

	fileURI := params.TextDocument.URI

	fileContents, ok := c.GetFileContents(fileURI)
	if !ok {
		// if the file contents is missing then we can't provide completions
		return nil, nil
	}

	lines := strings.Split(fileContents, "\n")
	if params.Position.Line >= uint(len(lines)) {
		return nil, nil
	}

	line := lines[params.Position.Line]

	if len(line) < int(params.Position.Character) || len(line) < 2 {
		return nil, nil
	}

	if !strings.Contains(line, " if ") && // if after if keyword
		!strings.Contains(line, " contains ") && // if after contains
		!strings.Contains(line, " else ") && // if after else
		!strings.Contains(line, "= ") && // if after assignment
		!strings.HasPrefix(line, "  ") { // if in rule body
		return nil, nil
	}

	words := strings.Split(line, " ")
	lastWord := words[len(words)-1]

	items := []types.CompletionItem{}

	for key, builtIn := range rego.BuiltIns {
		if strings.HasPrefix(key, lastWord) {
			items = append(items, types.CompletionItem{
				Label:  key,
				Kind:   3, // 3 is the kind for a function
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
