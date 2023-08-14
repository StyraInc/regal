package regal.ast_test

import future.keywords.if
import future.keywords.in

import data.regal.ast

test_find_vars if {
	policy := `
	package p

	import future.keywords

	global := "foo"

	allow if {
		a := global
		b := [c | c := input[x]] # can't capture x

		every d in input {
			d == "foo"
		}

		every e, f in input.bar {
			e == f
		}

		some g, h
		input.bar[g][h]
		some i in input
		some j, k in input

		[l, m, n] := [1, 2, 3]

		[o, [p, _]] := [1, [2, 1]]

		some _, [q, r] in [["foo", "bar"], [1, 2]]

		{"x": s} := {"x": 1}

		some [t] in [[1]]
	}
	`

	module := regal.parse_module("p.rego", policy)

	vars := ast.find_vars(module.rules)

	names := {name |
		some var in vars
		var.type == "var"
		name := var.value
	}

	names == {"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t"}
}

test_find_vars_comprehension_lhs if {
	policy := `
	package p

	import future.keywords

	allow if {
		a := [b | input[b]]
		c := {d | input[d]}
		e := {f: g | g := input[f]}
	}
	`

	module := regal.parse_module("p.rego", policy)

	vars := ast.find_vars(module.rules)

	names := {name |
		some var in vars
		var.type == "var"
		name := var.value
	}

	names == {"a", "b", "c", "d", "e", "f", "g"}
}

# https://github.com/StyraInc/regal/issues/168
test_function_decls_multiple_same_name if {
	policy := `package p

	import future.keywords.if

	f(x) := x if true
	f(y) := y if false
	`
	module := regal.parse_module("p.rego", policy)
	custom := ast.function_decls(module.rules)

	# we only need to assert there wasn't a conflict in the above
	# call, not what value was returned
	is_object(custom)
}

test_comment_blocks if {
	policy := `package p

# METADATA
# title: foo
# bar: invalid
allow := true

# not metadata

# another
# block
`

	module := regal.parse_module("p.rego", policy)
	blocks := ast.comment_blocks(module.comments)
	blocks == [
		[
			{"Location": {"col": 1, "file": "p.rego", "row": 3}, "Text": "IE1FVEFEQVRB"},
			{"Location": {"col": 1, "file": "p.rego", "row": 4}, "Text": "IHRpdGxlOiBmb28="},
			{"Location": {"col": 1, "file": "p.rego", "row": 5}, "Text": "IGJhcjogaW52YWxpZA=="},
		],
		[{"Location": {"col": 1, "file": "p.rego", "row": 8}, "Text": "IG5vdCBtZXRhZGF0YQ=="}],
		[
			{"Location": {"col": 1, "file": "p.rego", "row": 10}, "Text": "IGFub3RoZXI="},
			{"Location": {"col": 1, "file": "p.rego", "row": 11}, "Text": "IGJsb2Nr"},
		],
	]
}

test_find_vars_in_local_scope if {
	policy := `
	package p

	import future.keywords

	global := "foo"

	allow {
		a := global
		b := [c | c := input[x]] # can't capture x

		every d in input {
			d == "foo"
			e := "bar"
			e == "foo"
		}
	}`

	module := regal.parse_module("p.rego", policy)

	allow_rule := module.rules[1]

	var_locations := {
		"a": {"col": 3, "row": 9},
		"b": {"col": 3, "row": 10},
		"c": {"col": 13, "row": 10},
		"d": {"col": 9, "row": 12},
		"e": {"col": 4, "row": 14},
	}

	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.a)) == set()
	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.b)) == {"a"}
	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.c)) == {"a", "b", "c"}
	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.d)) == {"a", "b", "c"}
	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.e)) == {"a", "b", "c", "d"}
}

var_names(vars) := {var.value | some var in vars}

test_find_names_in_scope if {
	policy := `
	package p

	import future.keywords

	bar := "baz"

	global := "foo"

	comp := [foo | foo := input[_]]

	allow {
		a := global
		b := [c | c := input[x]] # can't capture x

		every d in input {
			d == "foo"
			e := "bar"
			e == "foo"
		}
	}`

	module := regal.parse_module("p.rego", policy)

	allow_rule := module.rules[3]

	in_scope := ast.find_names_in_scope(allow_rule, {"col": 1, "row": 30}) with input as module

	in_scope == {"bar", "global", "comp", "allow", "a", "b", "c", "d", "e"}
}
