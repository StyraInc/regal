# METADATA
# description: |
#   the `inputdotjson` provider returns suggestions based on the `input.json`
#   data structure (if such a file is found), so that e.g. content like:
#   ```json
#   {
#     "user": {"roles": ["admin"]},
#     "request": {"method": "GEt"}
#   }
#   ```
#   would suggest `input.user`, `input.user.roles`, `input.request`,
#   `input.request.method, and so on
package regal.lsp.completion.providers.inputdotjson

import rego.v1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

# METADATA
# description: items contains found suggestions from `input.json``
items contains item if {
	input.regal.context.input_dot_json_path

	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]
	word := location.ref_at(line, input.regal.context.location.col)

	some [suggestion, type] in _matching_input_suggestions

	item := {
		"label": suggestion,
		"kind": kind.variable,
		"detail": type,
		"documentation": {
			"kind": "markdown",
			"value": sprintf("(inferred from [`input.json`](%s))", [input.regal.context.input_dot_json_path]),
		},
		"textEdit": {
			"range": location.word_range(word, position),
			"newText": suggestion,
		},
	}
}

_matching_input_suggestions contains [suggestion, type] if {
	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	line != ""
	location.in_rule_body(line)

	word := location.ref_at(line, input.regal.context.location.col)

	some [suggestion, type] in _input_paths

	startswith(suggestion, word.text)
}

_input_paths contains [input_path, input_type] if {
	walk(input.regal.context.input_dot_json, [path, value])

	count(path) > 0

	# don't traverse into arrays
	every value in path {
		is_string(value)
	}

	input_type := type_name(value)
	input_path := concat(".", ["input", concat(".", path)])
}
