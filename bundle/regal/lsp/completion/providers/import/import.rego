package regal.lsp.completion.providers["import"]

import rego.v1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

items contains item if {
	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]
	word := location.word_at(line, input.regal.context.location.col)

	startswith("import", line)

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
