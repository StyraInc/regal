package regal.ast_test

import rego.v1

import data.regal.ast
import data.regal.capabilities

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

	vars := ast.find_vars(module.rules) with data.internal.combined_config as {"capabilities": capabilities.provided}

	names := {var.value |
		some var in vars
		var.type == "var"
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

	vars := ast.find_vars(module.rules) with data.internal.combined_config as {"capabilities": capabilities.provided}

	names := {var.value |
		some var in vars
		var.type == "var"
	}

	names == {"a", "b", "c", "d", "e", "f", "g"}
}

test_find_vars_function_ret_return_args if {
	policy := `
	package p

	import rego.v1

	allow if {
		walk(input, [path, value])
	}
	`

	module := regal.parse_module("p.rego", policy)

	vars := ast.find_vars(module.rules) with data.internal.combined_config as {"capabilities": capabilities.provided}
		with input.rules as []

	names := {var.value | some var in vars; var.type == "var"}
	names == {"path", "value"}
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

# regal ignore:rule-length
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
	blocks := ast.comment_blocks(module.comments) with input as module
	blocks == [
		[
			{"location": "3:1:5:15", "text": "IE1FVEFEQVRB"},
			{"location": "4:1:4:13", "text": "IHRpdGxlOiBmb28="},
			{"location": "5:1:5:15", "text": "IGJhcjogaW52YWxpZA=="},
		],
		[{"location": "8:1:8:15", "text": "IG5vdCBtZXRhZGF0YQ=="}],
		[
			{"location": "10:1:10:10", "text": "IGFub3RoZXI="},
			{"location": "11:1:11:8", "text": "IGJsb2Nr"},
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

	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.a)) with input as module == set()
	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.b)) with input as module == {"a"}
	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.c)) with input as module == {"a", "b", "c"}
	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.d)) with input as module == {"a", "b", "c", "d"}
	var_names(ast.find_vars_in_local_scope(allow_rule, var_locations.e)) with input as module == {"a", "b", "c", "d", "e"}
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

	ast.find_vars_in_local_scope(allow_rule, {"col": 10, "row": 10}) with input as module == [
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
		with data.internal.combined_config as {"capabilities": capabilities.provided}

	in_scope == {"bar", "global", "comp", "allow", "a", "b", "c", "d", "e"}
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

	{"x"} == ast.find_some_decl_names_in_scope(module.rules[0], {"col": 1, "row": 8}) with input as module
	{"x", "y", "z"} == ast.find_some_decl_names_in_scope(module.rules[0], {"col": 1, "row": 10}) with input as module
}

var_names(vars) := {var.value | some var in vars}

test_provided_capabilities_never_undefined if {
	capabilities.provided == {} with data.internal as {}
}

test_function_calls if {
	calls := ast.function_calls["0"] with input as ast.with_rego_v1(`
	rule if {
		x := 1
		f(2)
	}`)

	{"assign", "f"} == {call.name | some call in calls}
}

test_implicit_boolean_assignment if {
	ast.implicit_boolean_assignment(ast.with_rego_v1(`a.b if true`).rules[0])
}

test_ref_to_string if {
	ast.ref_to_string([{"type": "var", "value": "data"}]) == `data`
	ast.ref_to_string([{"type": "var", "value": "foo"}, {"type": "var", "value": "bar"}]) == `foo[bar]`
	ast.ref_to_string([{"type": "var", "value": "data"}, {"type": "string", "value": "/foo/"}]) == `data["/foo/"]`
	ast.ref_to_string([
		{"type": "var", "value": "foo"},
		{"type": "var", "value": "bar"},
		{"type": "var", "value": "baz"},
	]) == `foo[bar][baz]`
	ast.ref_to_string([
		{"type": "var", "value": "foo"},
		{"type": "var", "value": "bar"},
		{"type": "var", "value": "baz"},
		{"type": "string", "value": "qux"},
	]) == `foo[bar][baz].qux`
	ast.ref_to_string([
		{"type": "var", "value": "foo"},
		{"type": "string", "value": "~bar~"},
		{"type": "string", "value": "boo"},
		{"type": "var", "value": "baz"},
	]) == `foo["~bar~"].boo[baz]`
	ast.ref_to_string([
		{"type": "var", "value": "data"},
		{"type": "string", "value": "regal"},
		{"type": "string", "value": "lsp"},
		{"type": "string", "value": "completion_test"},
	]) == `data.regal.lsp.completion_test`
}

test_ref_static_to_string if {
	ast.ref_static_to_string([{"type": "var", "value": "data"}]) == `data`
	ast.ref_static_to_string([{"type": "var", "value": "foo"}, {"type": "var", "value": "bar"}]) == `foo`
	ast.ref_static_to_string([{"type": "var", "value": "data"}, {"type": "string", "value": "/foo/"}]) == `data["/foo/"]`
	ast.ref_static_to_string([
		{"type": "var", "value": "foo"},
		{"type": "string", "value": "bar"},
		{"type": "var", "value": "baz"},
	]) == `foo.bar`
	ast.ref_static_to_string([
		{"type": "var", "value": "foo"},
		{"type": "string", "value": "~bar~"},
		{"type": "string", "value": "qux"},
	]) == `foo["~bar~"].qux`
	ast.ref_static_to_string([
		{"type": "var", "value": "data"},
		{"type": "string", "value": "regal"},
		{"type": "string", "value": "lsp"},
		{"type": "string", "value": "completion_test"},
	]) == `data.regal.lsp.completion_test`
}

test_rule_head_locations if {
	policy := `package policy

import rego.v1

default allow := false

allow if true

reasons contains "foo"
reasons contains "bar"

default my_func(_) := false
my_func(1) := true

ref_rule[foo] := true if {
	some foo in [1,2,3]
}
`
	result := ast.rule_head_locations with input as regal.parse_module("p.rego", policy)

	result == {
		"data.policy.allow": {{"col": 9, "row": 5}, {"col": 1, "row": 7}},
		"data.policy.reasons": {{"col": 1, "row": 9}, {"col": 1, "row": 10}},
		"data.policy.my_func": {{"col": 9, "row": 12}, {"col": 1, "row": 13}},
		"data.policy.ref_rule": {{"col": 1, "row": 15}},
	}
}

test_public_rules_and_functions if {
	module := regal.parse_module("p.rego", `package p

foo := true

_bar := false

x.y := true

x._z := false
	`)

	public := ast.public_rules_and_functions with input as module

	{ast.ref_to_string(rule.head.ref) | some rule in public} == {"foo", "x.y"}
}
