// ComputeEdits is copied from https://github.com/kitagry/regols, the source repo's license is MIT and is copied below:
//
// MIT License
//
// # Copyright (c) 2023 Ryo Kitagawa
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
		return "", fmt.Errorf("failed to format Rego source file: %w", err)
	}

	return string(formatted), nil
}

// ComputeEdits computes diff edits from 2 string inputs.
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

// The opa fmt command does not allow configuration options, so we will have
// to ignore them if provided by the client. We can however log the warnings
// so that the caller has some chance to be made aware of why their options
// aren't applied.
func validateFormattingOptions(opts FormattingOptions) []string {
	warnings := make([]string, 0)

	if opts.InsertSpaces {
		warnings = append(warnings, "opa fmt: only tabs supported for indentation")
	}

	if !opts.TrimTrailingWhitespace {
		warnings = append(warnings, "opa fmt: trailing whitespace always trimmed")
	}

	if !opts.InsertFinalNewline {
		warnings = append(warnings, "opa fmt: final newline always inserted")
	}

	if !opts.TrimFinalNewlines {
		warnings = append(warnings, "opa fmt: final newlines always trimmed")
	}

	// opts.TabSize ignored as we don't support using spaces

	return warnings
}
