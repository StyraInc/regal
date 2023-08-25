package regal.ast

import future.keywords.contains
import future.keywords.every
import future.keywords.if
import future.keywords.in

import data.regal.opa

builtin_names := object.keys(opa.builtins)

# METADATA
# description: |
#   provide the package name / path as originally declared in the
#   input policy, so "package foo.bar" would return "foo.bar"
package_name := concat(".", [path.value |
	some i, path in input["package"].path
	i > 0
])

tests := [rule |
	some rule in input.rules
	not rule.head.args
	startswith(rule.head.name, "test_")
]

rule_names := {rule.head.name |
	some rule in input.rules
	not rule.head.args
}

# METADATA
# description: parse provided snippet with a generic package declaration added
policy(snippet) := regal.parse_module("policy.rego", concat("", [
	"package policy\n\n",
	snippet,
]))

# METADATA
# description: parses provided policy with all future keywords imported. Primarily for testing.
with_future_keywords(policy) := regal.parse_module("policy.rego", concat("", [
	`package policy

import future.keywords.contains
import future.keywords.every
import future.keywords.if
import future.keywords.in

`,
	policy,
]))

_find_nested_vars(obj) := [value |
	# regal ignore:function-arg-return
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
	# regal ignore:function-arg-return
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

_find_vars(_, value, _) := find_ref_vars(value) if {
	value.type == "ref"
}

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

_find_vars(_, value, _) := _find_object_comprehension_vars(value) if {
	value.type == "objectcomprehension"
}

find_some_decl_vars(rule) := [var |
	# regal ignore:function-arg-return
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
	# regal ignore:function-arg-return
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

_before_location(var, location) if {
	var.location.row < location.row
}

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

# METADATA
# description: |
#   traverses all nodes under provided node (using `walk`), and returns an array with
#   all calls to builtin functions
find_builtin_calls(node) := [value |
	# regal ignore:function-arg-return
	walk(node, [path, value])

	regal.last(path) == "terms"

	value[0].type == "ref"
	value[0].value[0].type == "var"
	value[0].value[0].value in builtin_names
]

# METADATA
# description: |
#   Returns custom functions declared in input policy in the same format as opa.builtins
function_decls(rules) := {name: args |
	some rule in rules
	rule.head.args

	name := rule.head.name

	# ensure we only get one set of args, or we'll have a conflict
	args := [{"args": [item |
		some arg in rule.head.args
		item := {"name": arg.value}
	]} |
		some rule in rules
		rule.head.name == name
	][0]
}

function_ret_in_args(fn_name, terms) if {
	rest := array.slice(terms, 1, count(terms))

	# for now, bail out of nested calls
	not "call" in {term.type | some term in rest}

	count(rest) > count(all_functions[fn_name].args)
}

# METADATA
# description: |
#   answers if provided rule is implicitly assigned boolean true, i.e. allow { .. } or not
implicit_boolean_assignment(rule) if {
	# note the missing location attribute here, which is how we distinguish
	# between implicit and explicit assignments
	rule.head.value == {"type": "boolean", "value": true}
}

all_functions := object.union(opa.builtins, function_decls(input.rules))

all_function_names := object.keys(all_functions)

comments_decoded := [decoded |
	some comment in input.comments
	decoded := object.union(comment, {"Text": base64.decode(comment.Text)})
]

comments["blocks"] := comment_blocks(comments_decoded)

comments["metadata_attributes"] := {
	"scope",
	"title",
	"description",
	"related_resources",
	"authors",
	"organizations",
	"schemas",
	"entrypoint",
	"custom",
}

comment_blocks(comments) := [partition |
	rows := [row |
		some comment in comments
		row := comment.Location.row
	]
	breaks := _splits(rows)

	some j, k in breaks
	partition := array.slice(
		comments,
		breaks[j - 1] + 1,
		k + 1,
	)
]

_splits(xs) := array.concat(
	array.concat(
		# [-1] ++ [ all indices where there's a step larger than one ] ++ [length of xs]
		# the -1 is because we're adding +1 in array.slice
		[-1],
		[i |
			some i in numbers.range(0, count(xs) - 1)
			xs[i + 1] != xs[i] + 1
		],
	),
	[count(xs)],
)
