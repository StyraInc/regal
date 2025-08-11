package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
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
