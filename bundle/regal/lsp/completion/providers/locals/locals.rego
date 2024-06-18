package regal.lsp.completion.providers.locals

import rego.v1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

items contains item if {
	position := location.to_position(input.regal.context.location)

	line := input.regal.file.lines[position.line]
	line != ""
	location.in_rule_body(line)

	word := location.word_at(line, input.regal.context.location.col)

	not excluded(line, position)

	some local in location.find_locals(input.rules, input.regal.context.location)

	startswith(local, word.text)

	item := {
		"label": local,
		"kind": kind.variable,
		"detail": "local variable",
		"textEdit": {
			"range": location.word_range(word, position),
			"newText": local,
		},
	}
}

# exclude local suggestions in function args definition,
# as those would recursively contribute to themselves
excluded(line, position) if _function_args_position(substring(line, 0, position.character))

_function_args_position(text) if {
	text == trim_left(text, " \t")
	contains(text, "(")
	not contains(text, "=")
}
