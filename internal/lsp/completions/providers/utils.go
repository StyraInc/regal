package providers

import (
	"regexp"
	"strings"

	"github.com/open-policy-agent/opa/v1/util"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
)

var (
	patternRuleBody   = regexp.MustCompile(`^\s+`)
	patternWhiteSpace = regexp.MustCompile(`\s+`)
)

// completionLineHelper returns the lines of a file and the current line for a given index. This
// function is used by multiple completion providers.
func completionLineHelper(c *cache.Cache, fileURI string, currentLineNumber uint) ([]string, string) {
	fileContents, ok := c.GetFileContents(fileURI)
	if !ok {
		return []string{}, ""
	}

	lines := strings.Split(fileContents, "\n")

	currentLine := ""
	if currentLineNumber < uint(len(lines)) {
		currentLine = lines[currentLineNumber]
	}

	return lines, currentLine
}

// groupKeyedRefsByDepth groups refs by their 'depth', where depth is the number of dots in the key.
// This is helpful when attempting to show shorter, higher level keys before longer, lower level keys.
func groupKeyedRefsByDepth(refs map[string]types.Ref) ([]int, map[int]map[string]types.Ref) {
	byDepth := make(map[int]map[string]types.Ref)

	for key, item := range refs {
		depth := strings.Count(key, ".")

		if _, ok := byDepth[depth]; !ok {
			byDepth[depth] = make(map[string]types.Ref)
		}

		byDepth[depth][key] = item
	}

	// items from higher depths should be shown first
	return util.KeysSorted(byDepth), byDepth
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
