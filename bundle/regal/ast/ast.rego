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

# METADATA
# description: |
#   provide the package name / path as originally declared in the
#   input policy, so "package foo.bar" would return "foo.bar"
package_name := concat(".", [path.value |
	some i, path in input["package"].path
	i > 0
])

# METADATA
# description: |
#   returns the name of the rule, i.e. "foo" for "foo { ... }"
#   note that what constitutes the "name" of a ref-head rule is
#   ambiguous at best, i.e. a["b"][c] { ... } ... in those cases
#   we currently return the first element of the ref, i.e. "a"
name(rule) := rule.head.ref[0].value

named_refs(refs) := [ref |
	some i, ref in refs
	_is_name(ref, i)
]

_is_name(ref, 0) if ref.type == "var"

_is_name(ref, pos) if {
	pos > 0
	ref.type == "string"
}

# allow := true, which expands to allow = true { true }
generated_body(rule) if rule.body[0].location == rule.head.value.location

generated_body(rule) if rule["default"] == true

# rule["message"] or
# rule contains "message"
generated_body(rule) if {
	rule.body[0].location.row == rule.head.key.location.row

	# this is a quirk in the AST — the generated body will have a location
	# set before the key, i.e. "message"
	rule.body[0].location.col < rule.head.key.location.col
}

# f("x")
generated_body(rule) if rule.body[0].location == rule.head.location

rules := [rule |
	some rule in input.rules
	not rule.head.args
]

tests := [rule |
	some rule in input.rules
	not rule.head.args
	startswith(name(rule), "test_")
]

functions := [rule |
	some rule in input.rules
	rule.head.args
]

function_arg_names(rule) := [arg.value | some arg in rule.head.args]

rule_and_function_names contains name(rule) if some rule in input.rules

rule_names contains name(rule) if some rule in rules

# METADATA
# description: parse provided snippet with a generic package declaration added
policy(snippet) := regal.parse_module("policy.rego", concat("", [
	"package policy\n\n",
	snippet,
]))

# METADATA
# description: parses provided policy with all future keywords imported. Primarily for testing.
with_rego_v1(policy) := regal.parse_module("policy.rego", concat("", [
	`package policy

import rego.v1

`,
	policy,
]))

_find_nested_vars(obj) := [value |
	walk(obj, [_, value])
	value.type == "var"
	indexof(value.value, "$") == -1
]

# simple assignment, i.e. `x := 100` returns `x`
# always returns a single var, but wrapped in an
# array for consistency
_find_assign_vars(_, value) := var if {
	value[1].type == "var"
	var := [value[1]]
}

# 'destructuring' array assignment, i.e.
# [a, b, c] := [1, 2, 3]
# or
# {a: b} := {"foo": "bar"}
_find_assign_vars(_, value) := var if {
	value[1].type in {"array", "object"}
	var := _find_nested_vars(value[1])
}

# var declared via `some`, i.e. `some x` or `some x, y`
_find_some_decl_vars(_, value) := [v |
	some v in value
	v.type == "var"
]

# single var declared via `some in`, i.e. `some x in y`
_find_some_in_decl_vars(_, value) := var if {
	arr := value[0].value
	count(arr) == 3

	var := _find_nested_vars(arr[1])
}

# two vars declared via `some in`, i.e. `some x, y in z`
_find_some_in_decl_vars(_, value) := var if {
	arr := value[0].value
	count(arr) == 4

	var := [v |
		some i in [1, 2]
		some v in _find_nested_vars(arr[i])
	]
}

# find vars like input[x].foo[y] where x and y are vars
# note: value.type == "ref" check must have been done before calling this function
find_ref_vars(value) := [var |
	some i, var in value.value

	# ignore first element, as it is the base ref (like input or data)
	i > 0
	var.type == "var"
]

# one or two vars declared via `every`, i.e. `every x in y {}`
# or `every`, i.e. `every x, y in y {}`
_find_every_vars(_, value) := var if {
	key_var := [v | v := value.key; v.type == "var"; indexof(v.value, "$") == -1]
	val_var := [v | v := value.value; v.type == "var"; indexof(v.value, "$") == -1]

	var := array.concat(key_var, val_var)
}

_find_term_vars(term) := [value |
	walk(term, [_, value])

	value.type == "var"
]

_find_set_or_array_comprehension_vars(value) := [value.value.term] if {
	value.value.term.type == "var"
} else := _find_term_vars(value.value.term)

_find_object_comprehension_vars(value) := array.concat(key, val) if {
	key := [value.value.key | value.value.key.type == "var"]
	val := [value.value.value | value.value.value.type == "var"]
}

_find_vars(path, value, last) := _find_assign_vars(path, value) if {
	last == "terms"
	value[0].type == "ref"
	value[0].value[0].type == "var"
	value[0].value[0].value == "assign"
}

_find_vars(_, value, _) := find_ref_vars(value) if value.type == "ref"

_find_vars(path, value, last) := _find_some_in_decl_vars(path, value) if {
	last == "symbols"
	value[0].type == "call"
}

_find_vars(path, value, last) := _find_some_decl_vars(path, value) if {
	last == "symbols"
	value[0].type != "call"
}

_find_vars(path, value, last) := _find_every_vars(path, value) if {
	last == "terms"
	value.domain
}

_find_vars(_, value, _) := _find_set_or_array_comprehension_vars(value) if {
	value.type in {"setcomprehension", "arraycomprehension"}
}

_find_vars(_, value, _) := _find_object_comprehension_vars(value) if value.type == "objectcomprehension"

find_some_decl_vars(rule) := [var |
	walk(rule, [path, value])

	regal.last(path) == "symbols"
	value[0].type != "call"

	some var in _find_some_decl_vars(path, value)
]

# METADATA
# description: |
#   traverses all nodes under provided node (using `walk`), and returns an array with
#   all variables declared via assignment (:=), `some`, `every` and in comprehensions
find_vars(node) := [var |
	walk(node, [path, value])

	some var in _find_vars(path, value, regal.last(path))
]

_function_arg_names(rule) := {arg.value |
	some arg in rule.head.args
	arg.type == "var"
}

# METADATA
# description: |
#   finds all vars declared in `rule` *before* the `location` provided
#   note: this isn't 100% accurate, as it doesn't take into account `=`
#   assignments / unification, but it's likely good enough since other rules
#   recommend against those
find_vars_in_local_scope(rule, location) := [var |
	some var in find_vars(rule)
	not startswith(var.value, "$")
	_before_location(var, location)
]

_before_location(var, location) if var.location.row < location.row

_before_location(var, location) if {
	var.location.row == location.row
	var.location.col < location.col
}

# METADATA
# description: |
#   similar to `find_vars_in_local_scope`, but returns all variable names in scope
#   of the given location *and* the rule names present in the scope (i.e. module)
find_names_in_scope(rule, location) := names if {
	fn_arg_names := _function_arg_names(rule)
	var_names := {var.value | some var in find_vars_in_local_scope(rule, location)}

	# parens below added by opa-fmt :)
	names := (rule_names | fn_arg_names) | var_names
}

# METADATA
# description: |
#   find all variables declared via `some` declarations (and *not* `some .. in`)
#   in the scope of the given location
find_some_decl_names_in_scope(rule, location) := {some_var.value |
	some some_var in find_some_decl_vars(rule)
	_before_location(some_var, location)
}

# METADATA
# description: |
#   determine if var in ref (e.g. `x` in `input[x]`) is used as input or output
is_output_var(rule, ref, location) if {
	startswith(ref.value, "$")
} else if {
	not ref.value in (find_names_in_scope(rule, location) - find_some_decl_names_in_scope(rule, location))
}

default is_ref(_) := false

is_ref(value) if value.type == "ref"

is_ref(value) if value[0].type == "ref"

all_refs contains value if {
	walk(input.rules, [_, value])

	is_ref(value)
}

all_refs contains value if {
	walk(input.imports, [_, value])

	is_ref(value)
}

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
static_rule_name(rule) := rule.head.ref[0].value if count(rule.head.ref) == 1

static_rule_name(rule) := concat(".", array.concat([rule.head.ref[0].value], [ref.value |
	some i, ref in rule.head.ref
	i > 0
])) if {
	count(rule.head.ref) > 1
	static_rule_ref(rule.head.ref)
}

# METADATA
# description: provides a set of all built-in function calls made in input policy
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
	some rule in functions

	rule_name := name(rule)

	# ensure we only get one set of args, or we'll have a conflict
	args := [[item |
		some arg in rule.head.args
		item := {"type": "any"}
	] |
		some rule in rules
		name(rule) == rule_name
	][0]

	decl := {"decl": {"args": args, "result": {"type": "any"}}}
}

function_ret_in_args(fn_name, terms) if {
	rest := array.slice(terms, 1, count(terms))

	# for now, bail out of nested calls
	not "call" in {term.type | some term in rest}

	count(rest) > count(all_functions[fn_name].decl.args)
}

# METADATA
# description: |
#   answers if provided rule is implicitly assigned boolean true, i.e. allow { .. } or not
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

all_functions := object.union(config.capabilities.builtins, function_decls(input.rules))

all_function_names := object.keys(all_functions)

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
