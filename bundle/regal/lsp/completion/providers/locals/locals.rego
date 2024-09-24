# METADATA
# description: provides completion suggestions for local symbols in scope
package regal.lsp.completion.providers.locals

import rego.v1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

# METADATA
# description: completion suggestions for local symbols
items contains item if {
	position := location.to_position(input.regal.context.location)

	line := input.regal.file.lines[position.line]
	line != ""

	location.in_rule_body(line)

	not _excluded(line, position)

	word := location.word_at(line, input.regal.context.location.col)
	parsed_current_file := data.workspace.parsed[input.regal.file.uri]

	some local in location.find_locals(
		parsed_current_file.rules,
		input.regal.context.location,
	)

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
_excluded(line, position) if _function_args_position(substring(line, 0, position.character))

_function_args_position(text) if {
	text == trim_left(text, " \t")
	contains(text, "(")
	not contains(text, "=")
}
