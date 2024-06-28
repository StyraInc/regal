package regal.ast

import rego.v1

import data.regal.config

scalar_types := {"boolean", "null", "number", "string"}

# regal ignore:external-reference
is_constant(value) if value.type in scalar_types

is_constant(value) if {
	value.type in {"array", "object"}
	not has_var(value)
}

has_var(value) if {
	walk(value.value, [_, term])
	term.type == "var"
}

builtin_names := object.keys(config.capabilities.builtins)

builtin_namespaces contains namespace if {
	some name in builtin_names
	namespace := split(name, ".")[0]
}

# METADATA
# description: |
#   provide the package name / path as originally declared in the
#   input policy, so "package foo.bar" would return "foo.bar"
package_name := concat(".", [path.value |
	some i, path in input["package"].path
	i > 0
])

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
# description: |
#   answers if the body was generated or not, i.e. not seen
#   in the original Rego file — for example `x := 1`
# scope: document

# METADATA
# description: covers case of allow := true, which expands to allow = true { true }
generated_body(rule) if rule.body[0].location == rule.head.value.location

# METADATA
# description: covers case of default rules
generated_body(rule) if rule["default"] == true

# METADATA
# description: covers case of rule["message"] or rule contains "message"
generated_body(rule) if {
	rule.body[0].location.row == rule.head.key.location.row

	# this is a quirk in the AST — the generated body will have a location
	# set before the key, i.e. "message"
	rule.body[0].location.col < rule.head.key.location.col
}

# METADATA
# description: covers case of f("x")
generated_body(rule) if rule.body[0].location == rule.head.location

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
# description: a list of the names for the giiven rule (if function)
function_arg_names(rule) := [arg.value | some arg in rule.head.args]

# METADATA
# description: all the rule and function names in the input AST
rule_and_function_names contains ref_to_string(rule.head.ref) if some rule in input.rules

# METADATA
# description: all identifers in the input AST (rule and functiin names, plus imported names)
identifiers := rule_and_function_names | imported_identifiers

# METADATA
# description: all rule names in the input AST (excluding functions)
rule_names contains ref_to_string(rule.head.ref) if some rule in rules

_function_arg_names(rule) := {arg.value |
	some arg in rule.head.args
	arg.type == "var"
}

# METADATA
# description: |
#   determine if var in ref (e.g. `x` in `input[x]`) is used as input or output
is_output_var(rule, ref, location) if {
	startswith(ref.value, "$")
} else if {
	not ref.value in (find_names_in_scope(rule, location) - find_some_decl_names_in_scope(rule, location))
}

# METADATA
# description: as the name implies, answers whether provided value is a ref
# scope: document
default is_ref(_) := false

is_ref(value) if value.type == "ref"

is_ref(value) if value[0].type == "ref"

all_rules_refs contains value if {
	walk(input.rules, [_, value])

	is_ref(value)
}

# METADATA
# title: all_refs
# description: set containing all references found in the input AST
# scope: document
all_refs contains value if some value in all_rules_refs

all_refs contains imported.path if some imported in input.imports

# METADATA
# title: ref_to_string
# description:  returns the "path" string of any given ref value
ref_to_string(ref) := concat(".", [_ref_part_to_string(i, part) | some i, part in ref])

_ref_part_to_string(0, ref) := ref.value

_ref_part_to_string(_, ref) := ref.value if ref.type == "string"

_ref_part_to_string(i, ref) := concat("", ["$", ref.value]) if {
	ref.type != "string"
	i > 0
}

static_ref(ref) if every t in array.slice(ref.value, 1, count(ref.value)) {
	t.type != "var"
}

static_rule_ref(ref) if every t in array.slice(ref, 1, count(ref)) {
	t.type != "var"
}

# METADATA
# description: |
#   return the name of a rule if, and only if it only has static parts with
#   no vars. This could be "username", or "user.name", but not "user[name]"
# scope: document
static_rule_name(rule) := rule.head.ref[0].value if count(rule.head.ref) == 1

static_rule_name(rule) := concat(".", array.concat([rule.head.ref[0].value], [ref.value |
	some i, ref in rule.head.ref
	i > 0
])) if {
	count(rule.head.ref) > 1
	static_rule_ref(rule.head.ref)
}

# METADATA
# description: provides a set of names of all built-in functions called in the input policy
builtin_functions_called contains name if {
	some value in all_refs

	value[0].value[0].type == "var"
	not value[0].value[0].value in {"input", "data"}

	name := concat(".", [value |
		some part in value[0].value
		value := part.value
	])

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
# description: |
#   answers if provided rule is implicitly assigned boolean true, i.e. allow { .. } or not
# scope: document
implicit_boolean_assignment(rule) if {
	# note the missing location attribute here, which is how we distinguish
	# between implicit and explicit assignments
	rule.head.value == {"type": "boolean", "value": true}
}

implicit_boolean_assignment(rule) if {
	# This handles the *quite* special case of
	# `a.b if true`, which is "rewritten" to `a.b = true` *and*  where a location is still added to the value
	# see https://github.com/open-policy-agent/opa/issues/6184 for details
	#
	# Do note that technically, it is possible to write a rule where the `true` value actually is on column 1, i.e.
	#
	# a.b =
	# true
	# if true
	#
	# If you write Rego like that — you're not going to use Regal anyway, are you? ¯\_(ツ)_/¯
	rule.head.value.type == "boolean"
	rule.head.value.value == true
	rule.head.value.location.col == 1
}

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
	row_text := lines[rule.head.location.row - 1]
	col_text := substring(row_text, rule.head.location.col - 1, -1)

	startswith(col_text, "{")
}
