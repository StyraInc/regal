# METADATA
# description: provides completion suggestions for the `import` keyword where applicable
package regal.lsp.completion.providers["import"]

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

# METADATA
# description: all completion suggestions for the import keyword
items contains item if {
	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	startswith("import", line)

	word := location.word_at(line, input.regal.context.location.col)

	item := {
		"label": "import",
		"kind": kind.keyword,
		"detail": "import <path>",
		"textEdit": {
			"range": location.word_range(word, position),
			"newText": "import ",
		},
	}
}
