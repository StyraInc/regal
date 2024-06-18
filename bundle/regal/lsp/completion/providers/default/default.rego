package regal.lsp.completion.providers["default"]

import rego.v1

import data.regal.ast

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

items contains item if {
	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	invoke_suggestion(line)

	item := {
		"label": "default",
		"kind": kind.keyword,
		"detail": "default <rule-name> := <value>",
		"textEdit": {
			"range": {
				"start": {
					"line": position.line,
					"character": 0,
				},
				"end": position,
			},
			"newText": "default ",
		},
	}
}

items contains item if {
	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	invoke_suggestion(line)

	some name in ast.rule_and_function_names

	item := {
		"label": sprintf("default %s := <value>", [name]),
		"kind": kind.keyword,
		"detail": sprintf("add default assignment for %s rule", [name]),
		"textEdit": {
			"range": {
				"start": {
					"line": position.line,
					"character": 0,
				},
				"end": position,
			},
			"newText": sprintf("default %s := ", [name]),
		},
	}
}

invoke_suggestion("")

invoke_suggestion(line) if startswith("default", line)
