package regal.ast

import rego.v1

import data.regal.config
import data.regal.util

scalar_types := {"boolean", "null", "number", "string"}

operators := {
	"and",
	"assign",
	"div",
	"eq",
	"equal",
	"gt",
	"gte",
	"internal.member_2",
	"internal.member_3",
	"lt",
	"lte",
	"minus",
	"mul",
	"neq",
	"or",
	"plus",
	"rem",
}

# regal ignore:external-reference
is_constant(value) if value.type in scalar_types

is_constant(value) if {
	value.type in {"array", "object"}
	not has_term_var(value.value)
}

builtin_names := object.keys(config.capabilities.builtins)

builtin_namespaces contains namespace if {
	some name in builtin_names
	namespace := split(name, ".")[0]
}

# METADATA
# description: |
#   provides the package path values (strings) as an array starting _from_ "data":
#   package foo.bar -> ["foo", "bar"]
package_path := [path.value |
	some i, path in input["package"].path
	i > 0
]

# METADATA
# description: |
#   provide the package name / path as originally declared in the
#   input policy, so "package foo.bar" would return "foo.bar"
package_name := concat(".", package_path)

named_refs(refs) := [ref |
	some i, ref in refs
	_is_name(ref, i)
]

_is_name(ref, 0) if ref.type == "var"

_is_name(ref, pos) if {
	pos > 0
	ref.type == "string"
}

# METADATA
# description: all the rules (excluding functions) in the input AST
rules := [rule |
	some rule in input.rules
	not rule.head.args
]

# METADATA
# description: all the test rules in the input AST
tests := [rule |
	some rule in input.rules
	not rule.head.args

	startswith(ref_to_string(rule.head.ref), "test_")
]

# METADATA
# description: all the functions declared in the input AST
functions := [rule |
	some rule in input.rules
	rule.head.args
]

# METADATA
# description: a list of the argument names for the given rule (if function)
function_arg_names(rule) := [arg.value | some arg in rule.head.args]

# METADATA
# description: all the rule and function names in the input AST
rule_and_function_names contains ref_to_string(rule.head.ref) if some rule in input.rules

# METADATA
# description: all identifiers in the input AST (rule and function names, plus imported names)
identifiers := rule_and_function_names | imported_identifiers

# METADATA
# description: all rule names in the input AST (excluding functions)
rule_names contains ref_to_string(rule.head.ref) if some rule in rules

# METADATA
# description: |
#   determine if var in var (e.g. `x` in `input[x]`) is used as input or output
# scope: document
is_output_var(rule, var) if {
	# test the cheap and common case first, and 'else' only when it's not
	startswith(var.value, "$")
} else if {
	not var.value in (rule_names | imported_identifiers) # regal ignore:external-reference

	num_above := sum([1 |
		some above in find_vars_in_local_scope(rule, var.location)
		above.value == var.value
	])
	num_some := sum([1 |
		some name in find_some_decl_names_in_scope(rule, var.location)
		name == var.value
	])

	# only the first ref variable in scope can be an output! meaning that:
	# allow if {
	#     some x
	#     input[x]    # <--- output
	#     data.bar[x] # <--- input
	# }
	num_above - num_some == 0
}

# METADATA
# description: as the name implies, answers whether provided value is a ref
# scope: document
is_ref(value) if value.type == "ref"

is_ref(value) if value[0].type == "ref"

# METADATA
# description: |
#   returns an array of all rule indices, as strings. this will be needed until
#   https://github.com/open-policy-agent/opa/issues/6736 is fixed
rule_index_strings := [s |
	some i, _ in _rules
	s := sprintf("%d", [i])
]

# METADATA
# description: |
#   a map containing all function calls (built-in and custom) in the input AST
#   keyed by rule index
function_calls[rule_index] contains call if {
	some rule_index in rule_index_strings
	some ref in found.refs[rule_index]

	name := ref_to_string(ref[0].value)
	args := [arg |
		some i, arg in array.slice(ref, 1, 100)

		not _exclude_arg(name, i, arg)
	]

	call := {
		"name": ref_to_string(ref[0].value),
		"location": ref[0].location,
		"args": args,
	}
}

# these will be aggregated as calls anyway, so let's try and keep this flat
_exclude_arg(_, _, arg) if arg.type == "call"

# first "arg" of assign is the variable to assign to.. special case we simply
# ignore here, as it's covered elsewhere
_exclude_arg("assign", 0, _)

# METADATA
# description: returns the "path" string of any given ref value
ref_to_string(ref) := concat(".", [_ref_part_to_string(i, part) | some i, part in ref])

_ref_part_to_string(0, ref) := ref.value

_ref_part_to_string(_, ref) := ref.value if ref.type == "string"

_ref_part_to_string(i, ref) := concat("", ["$", ref.value]) if {
	ref.type != "string"
	i > 0
}

# METADATA
# description: |
#   returns the string representation of a ref up until its first
#   non-static (i.e. variable) value, if any:
#   foo.bar -> foo.bar
#   foo.bar[baz] -> foo.bar
ref_static_to_string(ref) := ss if {
	rs := ref_to_string(ref)
	ss := substring(rs, 0, indexof(rs, ".$"))
}

static_ref(ref) if every t in array.slice(ref.value, 1, count(ref.value)) {
	t.type != "var"
}

# METADATA
# description: provides a set of names of all built-in functions called in the input policy
builtin_functions_called contains name if {
	name := function_calls[_][_].name
	name in builtin_names
}

# METADATA
# description: |
#   Returns custom functions declared in input policy in the same format as builtin capabilities
function_decls(rules) := {rule_name: decl |
	# regal ignore:external-reference
	some rule in functions

	rule_name := ref_to_string(rule.head.ref)

	# ensure we only get one set of args, or we'll have a conflict
	args := [[item |
		some arg in rule.head.args
		item := {"type": "any"}
	] |
		some rule in rules
		ref_to_string(rule.head.ref) == rule_name
	][0]

	decl := {"decl": {"args": args, "result": {"type": "any"}}}
}

function_ret_args(fn_name, terms) := array.slice(terms, count(all_functions[fn_name].decl.args) + 1, count(terms))

function_ret_in_args(fn_name, terms) if {
	rest := array.slice(terms, 1, count(terms))

	# for now, bail out of nested calls
	not "call" in {term.type | some term in rest}

	count(rest) > count(all_functions[fn_name].decl.args)
}

# METADATA
# description: answers if provided rule is implicitly assigned boolean true, i.e. allow { .. } or not
# scope: document
implicit_boolean_assignment(rule) if {
	# note the missing location attribute here, which is how we distinguish
	# between implicit and explicit assignments
	rule.head.value == {"type": "boolean", "value": true}
}

# or sometimes, like this...
implicit_boolean_assignment(rule) if rule.head.value.location == rule.head.location

# METADATA
# description: |
#   object containing all available built-in and custom functions in the
#   scope of the input AST, keyed by function name
all_functions := object.union(config.capabilities.builtins, function_decls(input.rules))

# METADATA
# description: |
#   set containing all available built-in and custom function names in the
#   scope of the input AST
all_function_names := object.keys(all_functions)

negated_expressions[rule] contains value if {
	some rule in input.rules

	walk(rule, [_, value])

	value.negated
}

# METADATA
# description: |
#   true if rule head contains no identifier, but is a chained rule body immediately following the previous one:
#   foo {
#       input.bar
#   } {	# <-- chained rule body
#       input.baz
#   }
is_chained_rule_body(rule, lines) if {
	head_loc := util.to_location_object(rule.head.location)

	row_text := lines[head_loc.row - 1]
	col_text := substring(row_text, head_loc.col - 1, -1)

	startswith(col_text, "{")
}

assignment_terms(expr) := [expr.terms[1], expr.terms[2]] if {
	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "assign"
}
