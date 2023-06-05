package regal.rules.style_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style

snake_case_violation := {
	"category": "style",
	"description": "Prefer snake_case for names",
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/prefer-snake-case", "style"),
	}],
	"title": "prefer-snake-case",
	"level": "error",
}

test_fail_camel_cased_rule_name if {
	report(`camelCase := 5`) == {object.union(
		snake_case_violation,
		{"location": {"col": 1, "file": "policy.rego", "row": 8, "text": `camelCase := 5`}},
	)}
}

test_success_snake_cased_rule_name if {
	report(`snake_case := 5`) == set()
}

test_fail_camel_cased_some_declaration if {
	report(`p {some fooBar; input[fooBar]}`) == {object.union(
		snake_case_violation,
		{"location": {"col": 9, "file": "policy.rego", "row": 8, "text": `p {some fooBar; input[fooBar]}`}},
	)}
}

test_success_snake_cased_some_declaration if {
	report(`p {some foo_bar; input[foo_bar]}`) == set()
}

test_fail_camel_cased_multiple_some_declaration if {
	report(`p {some x, foo_bar, fooBar; x = 1; foo_bar = 2; input[fooBar]}`) == {object.union(
		snake_case_violation,
		{"location": {
			"col": 21, "file": "policy.rego", "row": 8,
			"text": `p {some x, foo_bar, fooBar; x = 1; foo_bar = 2; input[fooBar]}`,
		}},
	)}
}

test_success_snake_cased_multiple_some_declaration if {
	report(`p {some x, foo_bar; x = 5; input[foo_bar]}`) == set()
}

test_fail_camel_cased_var_assignment if {
	report(`allow { camelCase := 5 }`) == {object.union(
		snake_case_violation,
		{"location": {"col": 9, "file": "policy.rego", "row": 8, "text": `allow { camelCase := 5 }`}},
	)}
}

test_fail_camel_cased_multiple_var_assignment if {
	report(`allow { snake_case := "foo"; camelCase := 5 }`) == {object.union(
		snake_case_violation,
		{"location": {
			"col": 30, "file": "policy.rego", "row": 8,
			"text": `allow { snake_case := "foo"; camelCase := 5 }`,
		}},
	)}
}

test_success_snake_cased_var_assignment if {
	report(`allow { snake_case := 5 }`) == set()
}

test_fail_camel_cased_some_in_value if {
	report(`allow { some cC in input }`) == {object.union(
		snake_case_violation,
		{"location": {"col": 14, "file": "policy.rego", "row": 8, "text": `allow { some cC in input }`}},
	)}
}

test_fail_camel_cased_some_in_key_value if {
	report(`allow { some cC, sc in input }`) == {object.union(
		snake_case_violation,
		{"location": {"col": 14, "file": "policy.rego", "row": 8, "text": `allow { some cC, sc in input }`}},
	)}
}

test_fail_camel_cased_some_in_key_value_2 if {
	report(`allow { some sc, cC in input }`) == {object.union(
		snake_case_violation,
		{"location": {"col": 18, "file": "policy.rego", "row": 8, "text": `allow { some sc, cC in input }`}},
	)}
}

test_success_snake_cased_some_in if {
	report(`allow { some sc in input }`) == set()
}

test_fail_camel_cased_every_value if {
	report(`allow { every cC in input { cC == 1 } }`) == {object.union(
		snake_case_violation,
		{"location": {"col": 15, "file": "policy.rego", "row": 8, "text": `allow { every cC in input { cC == 1 } }`}},
	)}
}

test_fail_camel_cased_every_key if {
	r := report(`allow { every cC, sc in input { cC == 1; sc == 2 } }`)
	r == {object.union(
		snake_case_violation,
		{"location": {
			"col": 15, "file": "policy.rego", "row": 8,
			"text": `allow { every cC, sc in input { cC == 1; sc == 2 } }`,
		}},
	)}
}

test_success_snake_cased_every if {
	report(`allow { every sc in input { sc == 1 } }`) == set()
}

# Prefer in operator over iteration

test_fail_use_in_operator_string_lhs if {
	r := report(`allow {
	"admin" == input.user.roles[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 13, "file": "policy.rego", "row": 9, "text": "\t\"admin\" == input.user.roles[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_number_lhs if {
	r := report(`allow {
	1 == input.lucky_numbers[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 7, "file": "policy.rego", "row": 9, "text": "\t1 == input.lucky_numbers[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_array_lhs if {
	r := report(`allow {
	[1] == input.arrays[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 9, "file": "policy.rego", "row": 9, "text": "\t[1] == input.arrays[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_boolean_lhs if {
	r := report(`allow {
	true == input.booleans[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 10, "file": "policy.rego", "row": 9, "text": "\ttrue == input.booleans[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_object_lhs if {
	r := report(`allow {
	{"x": "y"} == input.objects[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 16, "file": "policy.rego", "row": 9, "text": "\t{\"x\": \"y\"} == input.objects[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_null_lhs if {
	r := report(`allow {
	null == input.objects[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 10, "file": "policy.rego", "row": 9, "text": "\tnull == input.objects[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_set_lhs if {
	r := report(`allow {
	{"foo"} == input.objects[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 13, "file": "policy.rego", "row": 9, "text": "\t{\"foo\"} == input.objects[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_var_lhs if {
	report(`allow {
	admin == input.user.roles[_]
	}`) == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 11, "file": "policy.rego", "row": 9, "text": "\tadmin == input.user.roles[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_string_rhs if {
	report(`allow {
	input.user.roles[_] == "admin"
	}`) == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 2, "file": "policy.rego", "row": 9, "text": "\tinput.user.roles[_] == \"admin\""},
		"level": "error",
	}}
}

test_fail_use_in_operator_var_rhs if {
	r := report(`allow {
		input.user.roles[_] == admin
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 3, "file": "policy.rego", "row": 9, "text": "\t\tinput.user.roles[_] == admin"},
		"level": "error",
	}}
}

test_success_refs_both_sides if {
	report(`allow { required_roles[_] == input.user.roles[_] }`) == set()
}

test_success_uses_in_operator if {
	report(`allow { "admin" in input.user.roles[_] }`) == set()
}

# Line length

test_fail_line_too_long if {
	r := report(`allow {
foo == bar; bar == baz; [a, b, c, d, e, f] := [1, 2, 3, 4, 5, 6]; qux := [q | some q in input.nonsense]
	}`)
	r == {{
		"category": "style",
		"description": "Line too long",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/line-length", "style"),
		}],
		"title": "line-length",
		"location": {
			"col": 103, "file": "policy.rego", "row": 9,
			"text": `foo == bar; bar == baz; [a, b, c, d, e, f] := [1, 2, 3, 4, 5, 6]; qux := [q | some q in input.nonsense]`,
		},
		"level": "error",
	}}
}

test_success_line_not_too_long if {
	report(`allow { "foo" == "bar" }`) == set()
}

test_fail_unification_in_default_assignment if {
	report(`default x = false`) == {{
		"category": "style",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "style"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": "default x = false"},
		"level": "error",
	}}
}

test_success_assignment_in_default_assignment if {
	report(`default x := false`) == set()
}

test_fail_unification_in_object_rule_assignment if {
	r := report(`x["a"] = 1`)
	r == {{
		"category": "style",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "style"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": `x["a"] = 1`},
		"level": "error",
	}}
}

test_success_assignment_in_object_rule_assignment if {
	report(`x["a"] := 1`) == set()
}

# Some cases blocked by https://github.com/StyraInc/regal/issues/6 - e.g:
#
# allow = true { true }
#
# f(x) = 5

test_fail_todo_comment if {
	report(`# TODO: do someting clever`) == {{
		"category": "style",
		"description": "Avoid TODO comments",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/todo-comment", "style"),
		}],
		"title": "todo-comment",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": `# TODO: do someting clever`},
		"level": "error",
	}}
}

test_fail_fixme_comment if {
	report(`# fixme: this is broken`) == {{
		"category": "style",
		"description": "Avoid TODO comments",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/todo-comment", "style"),
		}],
		"title": "todo-comment",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": `# fixme: this is broken`},
		"level": "error",
	}}
}

test_success_no_todo_comment if {
	report(`# This code is great`) == set()
}

test_fail_function_references_input if {
	report(`f(_) { input.foo }`) == {{
		"category": "style",
		"description": "Reference to input, data or rule ref in function body",
		"location": {"col": 8, "file": "policy.rego", "row": 8, "text": `f(_) { input.foo }`},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "style"),
		}],
		"title": "external-reference",
		"level": "error",
	}}
}

test_fail_function_references_data if {
	report(`f(_) { data.foo }`) == {{
		"category": "style",
		"description": "Reference to input, data or rule ref in function body",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "style"),
		}],
		"title": "external-reference",
		"location": {"col": 8, "file": "policy.rego", "row": 8, "text": `f(_) { data.foo }`},
		"level": "error",
	}}
}

test_fail_function_references_rule if {
	r := report(`
foo := "bar"

f(x, y) {
	x == 5
	y == foo
}
	`)
	r == {{
		"category": "style",
		"description": "Reference to input, data or rule ref in function body",
		"location": {"col": 7, "file": "policy.rego", "row": 13, "text": `	y == foo`},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "style"),
		}],
		"title": "external-reference",
		"level": "error",
	}}
}

test_success_function_references_no_input_or_data if {
	report(`f(x) { x == true }`) == set()
}

test_success_function_references_no_input_or_data_reverse if {
	report(`f(x) { true == x }`) == set()
}

test_success_function_references_only_own_vars if {
	report(`f(x) { y := x; y == 10 }`) == set()
}

test_success_function_references_only_own_vars_nested if {
	report(`f(x, z) { y := x; y == [1, 2, z]}`) == set()
}

test_fail_rule_name_starts_with_get if {
	r := report(`get_foo := 1`)
	r == {{
		"category": "style",
		"description": "Avoid get_ and list_ prefix for rules and functions",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/avoid-get-and-list-prefix", "style"),
		}],
		"title": "avoid-get-and-list-prefix",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": "get_foo := 1"},
		"level": "error",
	}}
}

test_fail_function_name_starts_with_list if {
	r := report(`list_users(datasource) := ["we", "have", "no", "users"]`)
	r == {{
		"category": "style",
		"description": "Avoid get_ and list_ prefix for rules and functions",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/avoid-get-and-list-prefix", "style"),
		}],
		"title": "avoid-get-and-list-prefix",
		"location": {
			"col": 1, "file": "policy.rego", "row": 8,
			"text": `list_users(datasource) := ["we", "have", "no", "users"]`,
		},
		"level": "error",
	}}
}

test_fail_unconditional_assignment_in_body if {
	r := report(`x := y {
		y := 1
	}`)
	r == {{
		"category": "style",
		"description": "Unconditional assignment in rule body",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/unconditional-assignment", "style"),
		}],
		"title": "unconditional-assignment",
		"location": {"col": 3, "file": "policy.rego", "row": 9, "text": "\t\ty := 1"},
		"level": "error",
	}}
}

test_fail_unconditional_eq_in_body if {
	r := report(`x = y {
		y = 1
	}`)
	r == {{
		"category": "style",
		"description": "Unconditional assignment in rule body",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/unconditional-assignment", "style"),
		}],
		"title": "unconditional-assignment",
		"location": {"col": 3, "file": "policy.rego", "row": 9, "text": "\t\ty = 1"},
		"level": "error",
	}}
}

test_success_conditional_assignment_in_body if {
	report(`x := y { input.foo == "bar"; y := 1 }`) == set()
}

test_success_unconditional_assignment_but_with_in_body if {
	report(`x := y { y := 5 with input as 1 }`) == set()
}

test_fail_function_arg_return_value if {
	r := report(`foo := i { indexof("foo", "o", i) }`)
	r == {{
		"category": "style",
		"description": "Function argument used for return value",
		"level": "error",
		"location": {"col": 32, "file": "policy.rego", "row": 8, "text": "foo := i { indexof(\"foo\", \"o\", i) }"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/function-arg-return", "style"),
		}],
		"title": "function-arg-return",
	}}
}

test_success_function_arg_return_value_except_function if {
	r := style.report with input as ast.with_future_keywords(`foo := i { indexof("foo", "o", i) }`)
		with config.for_rule as {
			"level": "error",
			"except-functions": ["indexof"],
		}
	r == set()
}

report(snippet) := report if {
	# regal ignore:external-reference
	report := style.report with input as ast.with_future_keywords(snippet)
		with config.for_rule as {"level": "error", "max-line-length": 80}
}
