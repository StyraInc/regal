package regal.ast_test

import rego.v1

import data.regal.ast

# regal ignore:rule-length
test_find_vars if {
	policy := `
	package p

	import rego.v1

	global := "foo"

	allow if {
		a := global
		b := [c | c := input[d]]

		every e in input {
			e == "foo"
		}

		every f, g in input.bar {
			f == g
		}

		some h, i
		input.bar[h][i]
		some j in input
		some k, l in input

		[m, n, o] := [1, 2, 3]

		[p, [q, _]] := [1, [2, 1]]

		some _, [r, s] in [["foo", "bar"], [1, 2]]

		{"x": t} := {"x": 1}

		some [u] in [[1]]
	}
	`

	module := regal.parse_module("p.rego", policy)

	vars := ast.find_vars(module.rules)

	names := {name |
		some var in vars
		var.type == "var"
		name := var.value
	}

	names == {"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u"}
}

test_find_vars_comprehension_lhs if {
	policy := `
	package p

	import rego.v1

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

	import rego.v1

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
			{"Location": {"col": 1, "row": 3, "text": "IyBNRVRBREFUQQ=="}, "Text": "IE1FVEFEQVRB"},
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

# regal ignore:rule-length
test_find_vars_in_local_scope if {
	policy := `
	package p

	import rego.v1

	global := "foo"

	allow if {
		a := global
		b := [c | c := input[d]]

		every e in input {
			f == "foo"
			g := "bar"
			h == "foo"
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
	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.d)) == {"a", "b", "c", "d"}
	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.e)) == {"a", "b", "c", "d", "e"}
}

test_find_vars_in_local_scope_complex_comprehension_term if {
	policy := `
	package p

	import rego.v1

	allow if {
		a := [{"b": b} | c := input[b]]
	}`

	module := regal.parse_module("p.rego", policy)

	allow_rule := module.rules[0]

	ast.find_vars_in_local_scope(allow_rule, {"col": 10, "row": 10}) == [
		{"location": {"col": 3, "row": 7, "text": "YQ=="}, "type": "var", "value": "a"},
		{"location": {"col": 15, "row": 7, "text": "Yg=="}, "type": "var", "value": "b"},
		{"location": {"col": 20, "row": 7, "text": "Yw=="}, "type": "var", "value": "c"},
		{"location": {"col": 31, "row": 7, "text": "Yg=="}, "type": "var", "value": "b"},
	]
}

test_find_names_in_scope if {
	policy := `
	package p

	import rego.v1

	bar := "baz"

	global := "foo"

	comp := [foo | foo := input[_]]

	allow if {
		a := global
		b := [c | c := input[_]]

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

test_find_some_decl_vars if {
	policy := `
	package p

	import rego.v1

	allow if {
		foo := 1
		some x
		input[x]
		some y, z
		input[y][z] == x
	}`

	module := regal.parse_module("p.rego", policy)

	some_vars := ast.find_some_decl_vars(module.rules[0])

	var_names(some_vars) == {"x", "y", "z"}
}

test_find_some_decl_names_in_scope if {
	policy := `package p

	import rego.v1

	allow if {
		foo := 1
		some x
		input[x]
		some y, z
		input[y][z] == x
	}`

	module := regal.parse_module("p.rego", policy)

	ast.find_some_decl_names_in_scope(module.rules[0], {"col": 1, "row": 8}) == {"x"}
	ast.find_some_decl_names_in_scope(module.rules[0], {"col": 1, "row": 10}) == {"x", "y", "z"}
}

var_names(vars) := {var.value | some var in vars}

test_generated_body_function if {
	policy := `package p

	f("x")`

	module := regal.parse_module("p.rego", policy)

	ast.generated_body(module.rules[0])
}

test_all_refs if {
	policy := `package policy

	import data.foo.bar

    allow := data.foo.baz

    deny[message] {
		message := data.foo.bax
    }
    `

	module := regal.parse_module("p.rego", policy)

	r := ast.all_refs with input as module

	text_refs := {base64.decode(ref.location.text) | some ref in r}

	text_refs == {":=", "data.foo.bar", "data.foo.bax", "data.foo.baz"}
}

test_static_rule_name_one_part if {
	rule := {"head": {"ref": [{"type": "var", "value": "username"}]}}
	ast.static_rule_name(rule) == "username"
}

test_static_rule_name_multi_part if {
	rule := {"head": {"ref": [{"type": "var", "value": "user"}, {"type": "string", "value": "name"}]}}
	ast.static_rule_name(rule) == "user.name"
}

test_static_rule_name_var_part if {
	rule := {"head": {"ref": [{"type": "var", "value": "user"}, {"type": "var", "value": "name"}]}}
	not ast.static_rule_name(rule)
}
