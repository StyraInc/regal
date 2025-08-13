# METADATA
# description: package for signature help provider policies.
# schemas:
# - input: {type: object}
package regal.lsp["signature-help"]

import rego.v1

# METADATA
# description: main entry point for builtin signature help.
# entrypoint: true
default result := null

result := {
	"signatures": [{
		"label": _build_function_label(builtin_info.decl, func_info.name),
		# some builtins had a space at the start
		"documentation": trim_space(builtin_info.description),
		"parameters": _build_parameters(builtin_info.decl.args),
		"activeParameter": func_info.active_param - 1,
	}],
	"activeSignature": 0,
	"activeParameter": func_info.active_param - 1,
} if {
	func_info := _function_at_position(input.content, input.position)
	builtin_info := data.workspace.builtins[func_info.name]
}

default _function_at_position(_, _) := {}

_function_at_position(content, position) := {
	"name": last_match[1],
	"active_param": strings.count(last_match[2], ",") + 1,
} if {
	result := regex.find_all_string_submatch_n(
		`([a-zA-Z_][a-zA-Z0-9_.]*)\(([^)]*)$`,
		_text_up_to_position(split(content, "\n"), content, position),
		-1, # we want the last one specifically, so need to get all
	)

	last_match := result[count(result) - 1]
}

# when position is after the last line
_text_up_to_position(lines, content, position) := content if position.line >= count(lines)

# when char is off the last line
_text_up_to_position(lines, content, position) := content if {
	position.line < count(lines)
	current_line := lines[position.line]
	position.character >= count(current_line)
}

_text_up_to_position(lines, _, position) := concat("\n", all_lines) if {
	position.line < count(lines)

	current_line := lines[position.line]
	position.character < count(current_line)

	complete_lines := [line | some i, line in lines; i < position.line]
	partial_line := substring(current_line, 0, position.character)

	all_lines := array.concat(complete_lines, [partial_line])
}

_build_function_label(decl, func_name) := label if {
	param_labels := concat(", ", [_param_label(param) | some param in decl.args])
	label := sprintf("%s(%s) -> %s", [func_name, param_labels, decl.result.type])
}

_build_parameters(args) := [{
	"label": label,
	"documentation": sprintf("(%s): %s", [label, param.description]),
} |
	some param in args
	label := _param_label(param)
]

_param_label(param) := concat(": ", [param.name, param.type])
