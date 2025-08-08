package providers

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/hover"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

var (
	patternWhiteSpace = regexp.MustCompile(`\s+`)
	patternRuleBody   = regexp.MustCompile(`^\s+`)
)

type BuiltIns struct{}

func (*BuiltIns) Name() string {
	return "builtins"
}

func (*BuiltIns) Run(
	_ context.Context,
	c *cache.Cache,
	params types.CompletionParams,
	opts *Options,
) ([]types.CompletionItem, error) {
	if opts == nil {
		return nil, errors.New("builtins provider requires options")
	}

	fileURI := params.TextDocument.URI

	lines, currentLine := completionLineHelper(c, fileURI, params.Position.Line)

	if len(lines) < 1 || currentLine == "" || !inRuleBody(currentLine) {
		return []types.CompletionItem{}, nil
	}

	// default rules cannot contain calls
	if strings.HasPrefix(strings.TrimSpace(currentLine), "default ") {
		return []types.CompletionItem{}, nil
	}

	words := patternWhiteSpace.Split(strings.TrimSpace(currentLine), -1)
	lastWord := words[len(words)-1]
	items := make([]types.CompletionItem, 0, len(opts.Builtins))

	for _, builtIn := range opts.Builtins {
		key := builtIn.Name

		if builtIn.Infix != "" || builtIn.IsDeprecated() || !strings.HasPrefix(key, lastWord) {
			continue
		}

		items = append(items, types.CompletionItem{
			Label:  key,
			Kind:   completion.Function,
			Detail: "built-in function",
			Documentation: &types.MarkupContent{
				Kind:  "markdown",
				Value: hover.CreateHoverContent(builtIn),
			},
			TextEdit: &types.TextEdit{
				Range: types.RangeBetween(
					params.Position.Line,
					params.Position.Character-uint(len(lastWord)),
					params.Position.Line,
					params.Position.Character,
				),
				NewText: key,
			},
		})
	}

	return items, nil
}

// inRuleBody is a best-effort helper to determine if the current line is in a rule body.
func inRuleBody(currentLine string) bool {
	switch {
	case strings.Contains(currentLine, " if "):
		return true
	case strings.Contains(currentLine, " contains "):
		return true
	case strings.Contains(currentLine, " else "):
		return true
	case strings.Contains(currentLine, "= "):
		return true
	case patternRuleBody.MatchString(currentLine):
		return true
	}

	return false
}
