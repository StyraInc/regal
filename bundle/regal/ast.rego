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

# one or two vars declared via `every`, i.e. `every x in y {}`
# or `every`, i.e. `every x, y in y {}`
_find_every_vars(_, value) := var if {
	key_var := [v | v := value.key; v.type == "var"; indexof(v.value, "$") == -1]
	val_var := [v | v := value.value; v.type == "var"; indexof(v.value, "$") == -1]

	var := array.concat(key_var, val_var)
}

_find_vars(path, value, last) := _find_assign_vars(path, value) if {
	last == "terms"
	value[0].type == "ref"
	value[0].value[0].type == "var"
	value[0].value[0].value == "assign"
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

# METADATA
# description: |
#   traverses all nodes under provided node (using `walk`), and returns an array with
#   all variables declared via assignment (:=), `some` or `every`
find_vars(node) := [var |
	# regal ignore:function-arg-return
	walk(node, [path, value])

	some var in _find_vars(path, value, regal.last(path))
]

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
	not "call" in {t | t := rest[_].type}

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
