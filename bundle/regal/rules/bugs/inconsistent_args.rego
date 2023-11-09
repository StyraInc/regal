# METADATA
# description: Inconsistently named function arguments
package regal.rules.bugs["inconsistent-args"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	count(ast.functions) > 0

	# Comprehension indexing — as made obvious here it would be great
	# to have block level scoped ignore directives...
	function_args_by_name := {name: args_list |
		some i
		name := ast.ref_to_string(ast.functions[i].head.ref) # regal ignore:prefer-some-in-iteration
		args_list := [args |
			some j
			ast.ref_to_string(ast.functions[j].head.ref) == name # regal ignore:prefer-some-in-iteration
			args := ast.functions[j].head.args # regal ignore:prefer-some-in-iteration
		]
		count(args_list) > 1
	}

	some name, args_list in function_args_by_name

	# "Partition" the args by their position
	by_position := [s |
		some i, _ in args_list[0]
		s := [x | x := args_list[_][i]]
	]

	some position in by_position

	inconsistent_args(position)

	violation := result.fail(rego.metadata.chain(), result.location(find_function_by_name(name)))
}

inconsistent_args(position) if {
	named_vars := {arg.value |
		some arg in position
		arg.type == "var"
		not startswith(arg.value, "$")
	}
	count(named_vars) > 1
}

# Return the _second_ function found by name, as that
# is reasonably the location the inconsistency is found
find_function_by_name(name) := [fn |
	some fn in ast.functions
	ast.ref_to_string(fn.head.ref) == name
][1]
