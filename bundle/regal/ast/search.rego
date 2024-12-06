package regal.ast

import rego.v1

import data.regal.util

_find_nested_vars(obj) := [value |
	walk(obj, [_, value])
	value.type == "var"
	indexof(value.value, "$") == -1
]

# simple assignment, i.e. `x := 100` returns `x`
# always returns a single var, but wrapped in an
# array for consistency
_find_assign_vars(value) := var if {
	value[1].type == "var"
	var := [value[1]]
}

# 'destructuring' array assignment, i.e.
# [a, b, c] := [1, 2, 3]
# or
# {a: b} := {"foo": "bar"}
_find_assign_vars(value) := vars if {
	value[1].type in {"array", "object"}
	vars := _find_nested_vars(value[1])
}

# var declared via `some`, i.e. `some x` or `some x, y`
_find_some_decl_vars(value) := [v |
	some v in value
	v.type == "var"
]

# single var declared via `some in`, i.e. `some x in y`
_find_some_in_decl_vars(value) := vars if {
	arr := value[0].value
	count(arr) == 3

	vars := _find_nested_vars(arr[1])
}

# two vars declared via `some in`, i.e. `some x, y in z`
_find_some_in_decl_vars(value) := vars if {
	arr := value[0].value
	count(arr) == 4

	vars := [v |
		some i in [1, 2]
		some v in _find_nested_vars(arr[i])
	]
}

# METADATA
# description: |
#   find vars like input[x].foo[y] where x and y are vars
#   note: value.type == "ref" check must have been done before calling this function
find_ref_vars(value) := [var |
	some i, var in value.value

	i > 0
	var.type == "var"
]

# one or two vars declared via `every`, i.e. `every x in y {}`
# or `every`, i.e. `every x, y in y {}`
_find_every_vars(value) := vars if {
	key_var := [value.key |
		value.key.type == "var"
		indexof(value.key.value, "$") == -1
	]
	val_var := [value.value |
		value.value.type == "var"
		indexof(value.value.value, "$") == -1
	]

	vars := array.concat(key_var, val_var)
}

# METADATA
# description: |
#   traverses all nodes in provided terms (using `walk`), and returns an array with
#   all variables declared in terms, i,e [x, y] or {x: y}, etc.
find_term_vars(terms) := [term |
	walk(terms, [_, term])

	term.type == "var"
]

# METADATA
# description: |
#   traverses all nodes in provided terms (using `walk`), and returns true if any variable
#   is found in terms, with early exit (as opposed to find_term_vars)
has_term_var(terms) if {
	walk(terms, [_, term])

	term.type == "var"
}

_find_vars(value, last) := {"term": find_term_vars(function_ret_args(fn_name, value))} if {
	last == "terms"
	value[0].type == "ref"
	value[0].value[0].type == "var"
	value[0].value[0].value != "assign"

	fn_name := ref_to_string(value[0].value)

	not contains(fn_name, "$")
	fn_name in all_function_names # regal ignore:external-reference
	function_ret_in_args(fn_name, value)
}

# `=` isn't necessarily assignment, and only considering the variable on the
# left-hand side is equally dubious, but we'll treat `x = 1` as `x := 1` for
# the purpose of this function until we have a more robust way of dealing with
# unification
_find_vars(value, last) := {"assign": _find_assign_vars(value)} if {
	last == "terms"
	value[0].type == "ref"
	value[0].value[0].type == "var"
	value[0].value[0].value in {"assign", "eq"}
}

_find_vars(value, last) := {"somein": _find_some_in_decl_vars(value)} if {
	last == "symbols"
	value[0].type == "call"
}

_find_vars(value, last) := {"some": _find_some_decl_vars(value)} if {
	last == "symbols"
	value[0].type != "call"
}

_find_vars(value, last) := {"every": _find_every_vars(value)} if {
	last == "terms"
	value.domain
}

_find_vars(value, last) := {"args": arg_vars} if {
	last == "args"

	arg_vars := [arg |
		some arg in value
		arg.type == "var"
	]

	count(arg_vars) > 0
}

_rule_index(rule) := sprintf("%d", [i]) if {
	some i, r in _rules # regal ignore:external-reference
	r == rule
}

# METADATA
# description: |
#   traverses all nodes under provided node (using `walk`), and returns an array with
#   all variables declared via assignment (:=), `some`, `every` and in comprehensions
#   DEPRECATED: uses ast.found.vars instead
find_vars(node) := array.concat(
	[var |
		walk(node, [path, value])

		last := regal.last(path)
		last in {"terms", "symbols", "args"}

		var := _find_vars(value, last)[_][_]
	],
	[var |
		walk(node, [_, value])

		value.type == "ref"

		some x, var in value.value
		x > 0
		var.type == "var"
	],
)

# hack to work around the different input models of linting vs. the lsp package.. we
# should probably consider something more robust
_rules := input.rules

_rules := data.workspace.parsed[input.regal.file.uri].rules if not input.rules

# METADATA:
# description: |
#   object containing all variables found in the input AST, keyed first by the index of
#   the rule where the variables were found (as a numeric string), and then the context
#   of the variable, which will be one of:
#   - term
#   - assign
#   - every
#   - some
#   - somein
#   - ref
found.vars[rule_index][context] contains var if {
	some i, rule in _rules

	# converting to string until https://github.com/open-policy-agent/opa/issues/6736 is fixed
	rule_index := sprintf("%d", [i])

	walk(rule, [path, value])

	last := regal.last(path)
	last in {"terms", "symbols", "args"}

	some context, vars in _find_vars(value, last)
	some var in vars
}

found.vars[rule_index].ref contains var if {
	some i, rule in _rules

	# converting to string until https://github.com/open-policy-agent/opa/issues/6736 is fixed
	rule_index := sprintf("%d", [i])

	walk(rule, [_, value])

	value.type == "ref"

	some x, var in value.value
	x > 0
	var.type == "var"
}

# METADATA
# description: all refs found in module
# scope: document
found.refs[rule_index] contains value if {
	some i, rule in _rules

	# converting to string until https://github.com/open-policy-agent/opa/issues/6736 is fixed
	rule_index := sprintf("%d", [i])

	walk(rule, [_, value])

	value.type == "ref"
}

found.refs[rule_index] contains value if {
	some i, rule in _rules

	# converting to string until https://github.com/open-policy-agent/opa/issues/6736 is fixed
	rule_index := sprintf("%d", [i])

	walk(rule, [_, value])

	value[0].type == "ref"
}

# METADATA
# description: all symbols found in module
found.symbols[rule_index] contains value.symbols if {
	some i, rule in _rules

	# converting to string until https://github.com/open-policy-agent/opa/issues/6736 is fixed
	rule_index := sprintf("%d", [i])

	walk(rule, [_, value])
}

# METADATA
# description: all comprehensions found in module
found.comprehensions[rule_index] contains value if {
	some i, rule in _rules

	# converting to string until https://github.com/open-policy-agent/opa/issues/6736 is fixed
	rule_index := sprintf("%d", [i])

	walk(rule, [_, value])

	value.type in {"arraycomprehension", "objectcomprehension", "setcomprehension"}
}

# METADATA
# description: |
#   finds all vars declared in `rule` *before* the `location` provided
#   note: this isn't 100% accurate, as it doesn't take into account `=`
#   assignments / unification, but it's likely good enough since other rules
#   recommend against those
find_vars_in_local_scope(rule, location) := [var |
	var := found.vars[_rule_index(rule)][_][_] # regal ignore:external-reference

	not is_wildcard(var)
	_before_location(rule, var, util.to_location_object(location))
]

_end_location(location) := end if {
	loc := util.to_location_object(location)
	lines := split(loc.text, "\n")
	end := {
		"row": (loc.row + count(lines)) - 1,
		"col": loc.col + count(regal.last(lines)),
	}
}

# special case — the value location of the rule head "sees"
# all local variables declared in the rule body
_before_location(rule, _, location) if {
	loc := util.to_location_object(location)

	value_start := util.to_location_object(rule.head.value.location)

	loc.row >= value_start.row
	loc.col >= value_start.col

	value_end := _end_location(util.to_location_object(rule.head.value.location))

	loc.row <= value_end.row
	loc.col <= value_end.col
}

_before_location(_, var, location) if {
	util.to_location_object(var.location).row < util.to_location_object(location).row
}

_before_location(_, var, location) if {
	var_loc := util.to_location_object(var.location)
	loc := util.to_location_object(location)

	var_loc.row == loc.row
	var_loc.col < loc.col
}

# METADATA
# description: find *only* names in the local scope, and not e.g. rule names
find_names_in_local_scope(rule, location) := names if {
	fn_arg_names := _function_arg_names(rule)
	var_names := {var.value | some var in find_vars_in_local_scope(rule, util.to_location_object(location))}

	names := fn_arg_names | var_names
}

_function_arg_names(rule) := {arg.value |
	some arg in rule.head.args
	arg.type == "var"
}

# METADATA
# description: |
#   similar to `find_vars_in_local_scope`, but returns all variable names in scope
#   of the given location *and* the rule names present in the scope (i.e. module)
find_names_in_scope(rule, location) := names if {
	locals := find_names_in_local_scope(rule, util.to_location_object(location))

	# parens below added by opa-fmt :)
	names := (rule_names | imported_identifiers) | locals
}

# METADATA
# description: |
#   find all variables declared via `some` declarations (and *not* `some .. in`)
#   in the scope of the given location
find_some_decl_names_in_scope(rule, location) := {some_var.value |
	some some_var in found.vars[_rule_index(rule)]["some"] # regal ignore:external-reference
	_before_location(rule, some_var, location)
}

# METADATA
# description: all expressions in module
exprs[rule_index][expr_index] := expr if {
	some rule_index, rule in input.rules
	some expr_index, expr in rule.body
}
