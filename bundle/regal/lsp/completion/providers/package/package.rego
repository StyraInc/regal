# METADATA
# description: provides completion suggestions for the `package` keyword where applicable
package regal.lsp.completion.providers["package"]

import rego.v1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

# METADATA
# description: completion suggestions for package keyword
items contains item if {
	not strings.any_prefix_match(input.regal.file.lines, "package ")

	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	startswith("package", line)

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
