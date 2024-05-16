package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
)

type RegoV1 struct{}

func (*RegoV1) Run(c *cache.Cache, params types.CompletionParams, _ *Options) ([]types.CompletionItem, error) {
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

	if len(line) < int(params.Position.Character) {
		return nil, nil
	}

	// this completion provider applies on lines with import at the start
	if !strings.HasPrefix(line, "import ") {
		return nil, nil
	}

	words := strings.Split(line, " ")
	lastWord := words[len(words)-1]

	// We might be checking lines at this point like 'import r', 'import rego', 'import rego.v',
	// so here we take the last word (i.e. 'r', 'rego', 'rego.v') and check if that words is a
	// prefix of 'rego.v1'.
	//nolint:gocritic
	if !strings.HasPrefix("rego.v1", lastWord) {
		return nil, nil
	}

	return []types.CompletionItem{
		{
			Label:  "rego.v1",
			Kind:   9, // 9 is for Module
			Detail: "Use Rego v1",
			TextEdit: &types.TextEdit{
				Range: types.Range{
					Start: types.Position{
						Line:      params.Position.Line,
						Character: 7,
					},
					End: types.Position{
						Line:      params.Position.Line,
						Character: uint(len(line)),
					},
				},
				NewText: "rego.v1\n\n",
			},
		},
	}, nil
}
