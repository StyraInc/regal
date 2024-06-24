package regal.lsp.completion.providers["package"]

import rego.v1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

items contains item if {
	not strings.any_prefix_match(input.regal.file.lines, "package ")

	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	invoke_suggestion(line)

	item := {
		"label": "package",
		"kind": kind.keyword,
		"detail": "package <package-name>",
		"textEdit": {
			"range": location.from_start_of_line_to_position(position),
			"newText": "package ",
		},
	}
}

invoke_suggestion("")

# regal ignore:external-reference
invoke_suggestion(line) if startswith("package", line)
