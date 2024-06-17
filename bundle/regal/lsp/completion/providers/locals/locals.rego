package regal.lsp.completion.providers.locals

import rego.v1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

items contains item if {
	position := location.to_position(input.regal.context.location)

	line := input.regal.file.lines[position.line]
	line != ""
	location.in_rule_body(line)

	last_word := regal.last(regex.split(`\s+`, trim_space(line)))

	some local in location.find_locals(input.rules, input.regal.context.location)

	startswith(local, last_word)

	item := {
		"label": local,
		"kind": kind.variable,
		"detail": "local variable",
		"textEdit": {
			"range": {
				"start": {
					"line": position.line,
					"character": position.character - count(last_word),
				},
				"end": position,
			},
			"newText": local,
		},
	}
}
