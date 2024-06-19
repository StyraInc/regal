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

	invoke_suggestion(line, label)

	item := {
		"label": label,
		"kind": kind.snippet,
		"detail": "common name",
		"documentation": {
			"kind": "markdown",
			"value": sprintf("%q is a common rule name", [label]),
		},
		"textEdit": {
			"range": {
				"start": {
					"line": position.line,
					"character": 0,
				},
				"end": position,
			},
			"newText": sprintf("%s ", [label]),
		},
	}
}

invoke_suggestion("", _)

# regal ignore:external-reference
invoke_suggestion(line, label) if startswith(label, line)
