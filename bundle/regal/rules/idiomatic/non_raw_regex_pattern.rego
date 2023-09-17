# METADATA
# description: Use raw strings for regex patterns
package regal.rules.idiomatic["non-raw-regex-pattern"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

# Mapping of regex.* functions and the position(s)
# of their "pattern" attributes
re_pattern_functions := {
	"find_all_string_submatch_n": [1],
	"find_n": [1],
	"globs_match": [1, 2],
	"is_valid": [1],
	"match": [1],
	"replace": [2],
	"split": [1],
	"template_match": [1],
}

re_pattern_function_names := {
	"regex.find_all_string_submatch_n",
	"regex.find_n",
	"regex.globs_match",
	"regex.is_valid",
	"regex.match",
	"regex.replace",
	"regex.split",
	"regex.template_match",
}

any_regex_function_called if {
	some name in re_pattern_function_names
	name in ast.builtin_functions_called
}

report contains violation if {
	# skip expensive walk if no builtin regex function calls are registered
	any_regex_function_called

	walk(input.rules, [_, value])

	value[0].type == "ref"
	value[0].value[0].type == "var"
	value[0].value[0].value == "regex"

	# The name following "regex.", e.g. "match"
	name := value[0].value[1].value

	some pos in re_pattern_functions[name]

	value[pos].type == "string"

	loc := value[pos].location
	row := input.regal.file.lines[loc.row - 1]
	chr := substring(row, loc.col - 1, 1)

	chr == `"`

	violation := result.fail(rego.metadata.chain(), result.location(value[pos]))
}
