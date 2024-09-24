# METADATA
# description: Inconsistently named function arguments
package regal.rules.bugs["inconsistent-args"]

import rego.v1

import data.regal.ast
import data.regal.result
import data.regal.util

report contains violation if {
	count(ast.functions) > 0

	# Comprehension indexing — as made obvious here it would be great
	# to have block level scoped ignore directives...
	function_args_by_name := {name: args_list |
		some i
		name := ast.ref_to_string(ast.functions[i].head.ref)
		args_list := [args |
			some j
			ast.ref_to_string(ast.functions[j].head.ref) == name
			args := ast.functions[j].head.args
		]
		count(args_list) > 1
	}

	some name, args_list in function_args_by_name

	# "Partition" the args by their position
	by_position := [s |
		some i, _ in args_list[0]
		s := [item[i] | some item in args_list]
	]

	some position in by_position

	_inconsistent_args(position)

	violation := result.fail(rego.metadata.chain(), _args_location(_find_function_by_name(name)))
}

_args_location(fn) := loc if {
	# mostly to get the `text` attribute
	oloc := result.location(fn)

	farg := util.to_location_object(fn.head.args[0].location)
	larg := util.to_location_object(regal.last(fn.head.args).location)

	# use the location of the first and last arg for highlighting
	loc := object.union(oloc, {"location": {
		"row": farg.row,
		"col": farg.col,
		"end": {
			"row": larg.row,
			"col": larg.col + count(base64.decode(larg.text)),
		},
	}})
}

_inconsistent_args(position) if {
	named_vars := {arg.value |
		some arg in position
		arg.type == "var"
		not startswith(arg.value, "$")
	}
	count(named_vars) > 1
}

# Return the _second_ function found by name, as that
# is reasonably the location the inconsistency is found
_find_function_by_name(name) := [fn |
	some fn in ast.functions
	ast.ref_to_string(fn.head.ref) == name
][1]
