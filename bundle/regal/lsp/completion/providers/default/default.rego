# METADATA
# description: provides completion suggestions for the `default` keyword where applicable
package regal.lsp.completion.providers.default

import data.regal.ast

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

# METADATA
# description: all completion suggestions for default keyword
# scope: document
items contains item if {
	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	startswith("default", line)

	item := {
		"label": "default",
		"kind": kind.keyword,
		"detail": "default <rule-name> := <value>",
		"textEdit": {
			"range": location.from_start_of_line_to_position(position),
			"newText": "default ",
		},
	}
}

items contains item if {
	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	startswith("default", line)

	some name in ast.rule_and_function_names

	item := {
		"label": sprintf("default %s := <value>", [name]),
		"kind": kind.keyword,
		"detail": sprintf("add default assignment for %s rule", [name]),
		"textEdit": {
			"range": location.from_start_of_line_to_position(position),
			"newText": sprintf("default %s := ", [name]),
		},
	}
}
