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
		"ref": "https://docs.styra.com/regal/rules/prefer-snake-case",
	}],
	"title": "prefer-snake-case",
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
	report(`allow { every cC, sc in input { cC == 1; print(sc) } }`) == {object.union(
		snake_case_violation,
		{"location": {
			"col": 15, "file": "policy.rego", "row": 8,
			"text": `allow { every cC, sc in input { cC == 1; print(sc) } }`,
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
			"ref": "https://docs.styra.com/regal/rules/use-in-operator",
		}],
		"title": "use-in-operator",
		"location": {"col": 13, "file": "policy.rego", "row": 9, "text": "\t\"admin\" == input.user.roles[_]"},
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
			"ref": "https://docs.styra.com/regal/rules/use-in-operator",
		}],
		"title": "use-in-operator",
		"location": {"col": 7, "file": "policy.rego", "row": 9, "text": "\t1 == input.lucky_numbers[_]"},
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
			"ref": "https://docs.styra.com/regal/rules/use-in-operator",
		}],
		"title": "use-in-operator",
		"location": {"col": 9, "file": "policy.rego", "row": 9, "text": "\t[1] == input.arrays[_]"},
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
			"ref": "https://docs.styra.com/regal/rules/use-in-operator",
		}],
		"title": "use-in-operator",
		"location": {"col": 10, "file": "policy.rego", "row": 9, "text": "\ttrue == input.booleans[_]"},
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
			"ref": "https://docs.styra.com/regal/rules/use-in-operator",
		}],
		"title": "use-in-operator",
		"location": {"col": 16, "file": "policy.rego", "row": 9, "text": "\t{\"x\": \"y\"} == input.objects[_]"},
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
			"ref": "https://docs.styra.com/regal/rules/use-in-operator",
		}],
		"title": "use-in-operator",
		"location": {"col": 10, "file": "policy.rego", "row": 9, "text": "\tnull == input.objects[_]"},
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
			"ref": "https://docs.styra.com/regal/rules/use-in-operator",
		}],
		"title": "use-in-operator",
		"location": {"col": 13, "file": "policy.rego", "row": 9, "text": "\t{\"foo\"} == input.objects[_]"},
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
			"ref": "https://docs.styra.com/regal/rules/use-in-operator",
		}],
		"title": "use-in-operator",
		"location": {"col": 11, "file": "policy.rego", "row": 9, "text": "\tadmin == input.user.roles[_]"},
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
			"ref": "https://docs.styra.com/regal/rules/use-in-operator",
		}],
		"title": "use-in-operator",
		"location": {"col": 2, "file": "policy.rego", "row": 9, "text": "\tinput.user.roles[_] == \"admin\""},
	}}
}

test_fail_use_in_operator_var_rhs if {
	report(`allow {
		input.user.roles[_] == admin
	}`) == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/use-in-operator",
		}],
		"title": "use-in-operator",
		"location": {"col": 3, "file": "policy.rego", "row": 9, "text": "\t\tinput.user.roles[_] == admin"},
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
			"ref": "https://docs.styra.com/regal/rules/line-length",
		}],
		"title": "line-length",
		"location": {
			"col": 103, "file": "policy.rego", "row": 9,
			"text": `foo == bar; bar == baz; [a, b, c, d, e, f] := [1, 2, 3, 4, 5, 6]; qux := [q | some q in input.nonsense]`,
		},
	}}
}

test_success_line_not_too_long if {
	report(`allow { "foo" == "bar" }`) == set()
}

report(snippet) := report if {
	# regal ignore:input-or-data-reference
	report := style.report with input as ast.with_future_keywords(snippet)
		with config.for_rule as {"enabled": true, "max-line-length": 80}
}
