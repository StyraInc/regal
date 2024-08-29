package regal.rules.idiomatic["prefer-set-or-object-rule_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.idiomatic["prefer-set-or-object-rule"] as rule

test_fail_set_comprehension_could_be_rule if {
	module := ast.with_rego_v1(`my_set := {s |
		some s in input
		s > 10
	}`)

	r := rule.report with input as module
	r == {{
		"category": "idiomatic",
		"description": "Prefer set or object rule over comprehension",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 5, "text": "my_set := {s |"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/prefer-set-or-object-rule", "idiomatic"),
		}],
		"title": "prefer-set-or-object-rule",
	}}
}

test_fail_object_comprehension_could_be_rule if {
	module := ast.with_rego_v1(`my_obj := {k: v |
		some k, v in input
		v == "foo"
	}`)

	r := rule.report with input as module
	r == {{
		"category": "idiomatic",
		"description": "Prefer set or object rule over comprehension",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 5, "text": "my_obj := {k: v |"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/prefer-set-or-object-rule", "idiomatic"),
		}],
		"title": "prefer-set-or-object-rule",
	}}
}

test_success_set_comprehension_array_to_set_conversion_ref_iteration if {
	module := ast.with_rego_v1(`my_set := {s | s := arr[_]}`)

	r := rule.report with input as module
	r == set()
}

test_success_set_comprehension_array_to_set_conversion_ref_nested_iteration if {
	module := ast.with_rego_v1(`my_set := {s | s := a.b.c[_]}`)

	r := rule.report with input as module
	r == set()
}

test_success_set_comprehension_array_to_set_conversion_ref_nested_iteration_sub_attribute if {
	module := ast.with_rego_v1(`my_set := {s | s := a.b.c[_].d}`)

	r := rule.report with input as module
	r == set()
}

test_success_set_comprehension_array_to_set_conversion_some_in if {
	module := ast.with_rego_v1(`my_set := {s | some s in arr}`)

	r := rule.report with input as module
	r == set()
}

test_success_set_comprehension_but_rule_body if {
	module := ast.with_rego_v1(`my_set := {s | some s in arr; s == ""} if { some_condition }`)

	r := rule.report with input as module
	r == set()
}
