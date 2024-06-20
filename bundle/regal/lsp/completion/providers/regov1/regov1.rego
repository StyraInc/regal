package regal.lsp.completion.providers.regov1

import rego.v1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

items contains item if {
	not strings.any_prefix_match(input.regal.file.lines, "import rego.v1")

	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	startswith(line, "import ")

	word := location.ref_at(line, input.regal.context.location.col)

	invoke_suggestion(word)

	item := {
		"label": "rego.v1",
		"kind": kind.module,
		"detail": "use rego.v1",
		"textEdit": {
			"range": location.word_range(word, position),
			"newText": "rego.v1\n\n",
		},
	}
}

invoke_suggestion(word) if {
	startswith("rego.v1", word.text)
}
