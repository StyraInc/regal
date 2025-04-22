# METADATA
# description: Function argument can be narrowed
package regal.rules.custom["narrow-argument"]

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

report contains violation if {
	some name, arg
	refs := _args_refs[name][arg]

	narrowed := _narrow(refs)
	indices := _args_indices[name]

	not _arg_used_in_call(indices, arg)

	location := result.location(_first_named_arg_location(indices, arg))

	violation := result.fail(rego.metadata.chain(), object.union(
		location,
		{"description": _message(count(refs), arg, narrowed)},
	))
}

_message(1, arg, narrowed) := sprintf(
	"Argument %s only referenced as %s, value passed can be narrowed",
	[arg, narrowed],
)

_message(n, arg, narrowed) := sprintf(
	"Argument %s always referenced by a common prefix, value passed can be narrowed to %s",
	[arg, narrowed],
) if {
	n > 1
}

_narrow(refs) := narrowed if {
	count(refs) == 1

	arr := util.any_set_item(refs)

	count(arr) > 1

	narrowed := concat(".", arr)
}

_narrow(refs) := narrowed if {
	count(refs) > 1

	prefix := util.longest_prefix(refs)

	count(prefix) > 1

	narrowed := concat(".", prefix)
}

_first_named_arg_location(indices, name) := [arg.location |
	some rule_index in indices
	some arg in input.rules[rule_index].head.args

	arg.type == "var"
	arg.value == name
][0]

_arg_used_in_call(indices, name) if {
	some i in indices
	some call in ast.function_calls[ast.rule_index_strings[i]]
	some arg in call.args

	# only check for vars here, as refs are already dealt with
	arg.type == "var"
	arg.value == name
}

_args_refs[name][arg] contains ref if {
	some name, arg
	ref := _functions[name][_].args_refs[arg][_]
}

_args_indices[name] contains rule_index if {
	some name
	rule_index := _functions[name][_].rule_index
}

_functions[name] contains func if {
	cfg := config.rules.custom["narrow-argument"]

	some i
	args := input.rules[i].head.args

	variable_args := {arg.value |
		some arg in args
		arg.type == "var"
		not startswith(arg.value, "$")
		not _exclude_arg(cfg, arg.value)
	}

	# we don't care for functions without named variable arguments
	count(variable_args) > 0

	args_refs := {arg: ref_vals |
		arg := ast.found.refs[ast.rule_index_strings[i]][_].value[0].value
		arg in variable_args
		ref_vals := {vals |
			some g
			ast.found.refs[ast.rule_index_strings[i]][g].value[0].value == arg

			ref := ast.found.refs[ast.rule_index_strings[i]][g].value
			vals := [part.value |
				some part in array.slice(ref, 0, _first_var_pos(ref))
			]
		}
	}

	name := ast.ref_to_string(input.rules[i].head.ref)
	func := {"rule_index": i, "args_refs": args_refs}
}

_first_var_pos(ref) := pos if {
	pos := [i |
		some i, part in ref
		part.type == "var"
		i > 0
	][0]
} else := count(ref) + 1

_exclude_arg(cfg, name) if {
	name in cfg["exclude-args"]
}
