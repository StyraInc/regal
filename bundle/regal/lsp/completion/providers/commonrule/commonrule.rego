package regal.lsp.completion.providers.commonrule

import rego.v1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

suggested_names := {
	"allow",
	"authorized",
	"deny",
}

items contains item if {
	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	some label in suggested_names

	startswith(label, line)

	item := {
		"label": label,
		"kind": kind.snippet,
		"detail": "common name",
		"documentation": {
			"kind": "markdown",
			"value": sprintf("%q is a common rule name", [label]),
		},
		"textEdit": {
			"range": location.from_start_of_line_to_position(position),
			"newText": sprintf("%s ", [label]),
		},
	}
}
