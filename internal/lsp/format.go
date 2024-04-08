package lsp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/format"
)

func Format(path, contents string, opts format.Opts) (string, error) {
	formatted, err := format.SourceWithOpts(filepath.Base(path), []byte(contents), opts)
	if err != nil {
		return "", fmt.Errorf("failed to format Rego source file: %v", err)
	}

	return string(formatted), nil
}

// ComputeEdits computes diff edits from 2 string inputs
func ComputeEdits(before, after string) []TextEdit {
	ops := operations(splitLines(before), splitLines(after))
	edits := make([]TextEdit, 0, len(ops))
	for _, op := range ops {
		switch op.Kind {
		case Delete:
			// Delete: unformatted[i1:i2] is deleted.
			edits = append(edits, TextEdit{Range: Range{
				Start: Position{Line: op.I1, Character: 0},
				End:   Position{Line: op.I2, Character: 0},
			}})
		case Insert:
			// Insert: formatted[j1:j2] is inserted at unformatted[i1:i1].
			if content := strings.Join(op.Content, ""); content != "" {
				edits = append(edits, TextEdit{
					Range: Range{
						Start: Position{Line: op.I1, Character: 0},
						End:   Position{Line: op.I2, Character: 0},
					},
					NewText: content,
				})
			}
		case Equal:
		}
	}
	return edits
}
