package regal.lsp.completion.providers.booleans

import rego.v1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

items contains item if {
	position := location.to_position(input.regal.context.location)

	line := input.regal.file.lines[position.line]
	line != ""

	words := regex.split(`\s+`, line)

	words_on_line := count(words)

	previous_word := words[words_on_line - 2]

	previous_word in {"==", ":="}

	word := location.word_at(line, input.regal.context.location.col)

	some b in ["true", "false"]

	startswith(b, word.text)

	item := {
		"label": b,
		"kind": kind.constant,
		"detail": "boolean value",
		"textEdit": {
			"range": location.word_range(word, position),
			"newText": b,
		},
	}
}
