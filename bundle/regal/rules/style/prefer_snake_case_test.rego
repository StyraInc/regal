package regal.rules.style["prefer-snake-case_test"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.rules.style["prefer-snake-case"] as rule

test_fail_camel_cased_rule_name if {
	r := rule.report with input as ast.policy(`camelCase := 5`)
	r == expected_with_locations([{"col": 1, "file": "policy.rego", "row": 3, "text": `camelCase := 5`}])
}

test_success_snake_cased_rule_name if {
	r := rule.report with input as ast.policy(`snake_case := 5`)
	r == set()
}

test_fail_camel_cased_some_declaration if {
	r := rule.report with input as ast.policy(`p {some fooBar; input[_]}`)
	r == expected_with_locations([{"col": 9, "file": "policy.rego", "row": 3, "text": `p {some fooBar; input[_]}`}])
}

test_success_snake_cased_some_declaration if {
	r := rule.report with input as ast.policy(`p {some foo_bar; input[foo_bar]}`)
	r == set()
}

test_fail_camel_cased_multiple_some_declaration if {
	r := rule.report with input as ast.with_rego_v1(`p if {
		some x, foo_bar, fooBar; x = 1; foo_bar = 2; input[_]
	}`)
	r == expected_with_locations([{
		"col": 20,
		"file": "policy.rego",
		"row": 6,
		"text": "\t\tsome x, foo_bar, fooBar; x = 1; foo_bar = 2; input[_]",
	}])
}

test_success_snake_cased_multiple_some_declaration if {
	r := rule.report with input as ast.policy(`p {some x, foo_bar; x = 5; input[foo_bar]}`)
	r == set()
}

test_fail_camel_cased_var_assignment if {
	r := rule.report with input as ast.policy(`allow { camelCase := 5 }`)
	r == expected_with_locations([{"col": 9, "file": "policy.rego", "row": 3, "text": `allow { camelCase := 5 }`}])
}

test_fail_camel_cased_multiple_var_assignment if {
	r := rule.report with input as ast.policy(`allow { snake_case := "foo"; camelCase := 5 }`)
	r == expected_with_locations([{
		"col": 30,
		"file": "policy.rego",
		"row": 3,
		"text": `allow { snake_case := "foo"; camelCase := 5 }`,
	}])
}

test_success_snake_cased_var_assignment if {
	r := rule.report with input as ast.policy(`allow { snake_case := 5 }`)
	r == set()
}

test_fail_camel_cased_some_in_value if {
	r := rule.report with input as ast.with_rego_v1(`allow if { some cC in input }`)
	r == expected_with_locations([{
		"col": 17,
		"file": "policy.rego",
		"row": 5,
		"text": `allow if { some cC in input }`,
	}])
}

test_fail_camel_cased_some_in_key_value if {
	r := rule.report with input as ast.with_rego_v1(`allow if { some cC, sc in input }`)
	r == expected_with_locations([{
		"col": 17,
		"file": "policy.rego",
		"row": 5,
		"text": `allow if { some cC, sc in input }`,
	}])
}

test_fail_camel_cased_some_in_key_value_2 if {
	r := rule.report with input as ast.with_rego_v1(`allow if { some sc, cC in input }`)
	r == expected_with_locations([{
		"col": 21,
		"file": "policy.rego",
		"row": 5,
		"text": `allow if { some sc, cC in input }`,
	}])
}

test_success_snake_cased_some_in if {
	r := rule.report with input as ast.with_rego_v1(`allow if { some sc in input }`)
	r == set()
}

test_fail_camel_cased_every_value if {
	r := rule.report with input as ast.with_rego_v1(`allow if { every cC in input { cC == 1 } }`)
	r == expected_with_locations([{
		"col": 18,
		"file": "policy.rego",
		"row": 5,
		"text": `allow if { every cC in input { cC == 1 } }`,
	}])
}

test_fail_camel_cased_every_key if {
	r := rule.report with input as ast.with_rego_v1(`allow if { every cC, sc in input { cC == 1; sc == 2 } }`)
	r == expected_with_locations([{
		"col": 18, "file": "policy.rego", "row": 5,
		"text": `allow if { every cC, sc in input { cC == 1; sc == 2 } }`,
	}])
}

test_success_snake_cased_every if {
	r := rule.report with input as ast.with_rego_v1(`allow if { every sc in input { sc == 1 } }`)
	r == set()
}

# https://github.com/open-policy-agent/opa/issues/6860
test_fail_location_provided_even_when_not_in_ref if {
	r := rule.report with input as ast.with_rego_v1(`foo.Bar := true`)
	r == expected_with_locations([{"col": 5, "file": "policy.rego", "row": 5, "text": "foo.Bar := true"}])
}

expected_with_locations(locations) := {with_location |
	expected := {
		"category": "style",
		"description": "Prefer snake_case for names",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/prefer-snake-case", "style"),
		}],
		"title": "prefer-snake-case",
		"level": "error",
	}

	some location in locations
	with_location := object.union(expected, {"location": location})
}
