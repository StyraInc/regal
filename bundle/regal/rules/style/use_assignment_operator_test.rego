package regal.rules.style["use-assignment-operator_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style["use-assignment-operator"] as rule

test_fail_unification_in_regular_assignment if {
	r := rule.report with input as ast.policy(`foo = false`)
	r == {{
		"category": "style",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "style"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "foo = false"},
		"level": "error",
	}}
}

test_fail_not_implicit_boolean_assignment_with_body if {
	r := rule.report with input as ast.policy(`allow = true { true }`)
	r == {{
		"category": "style",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "style"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "allow = true { true }"},
		"level": "error",
	}}
}

test_fail_not_implicit_boolean_assignment if {
	r := rule.report with input as ast.policy(`foo = true`)
	r == {{
		"category": "style",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "style"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "foo = true"},
		"level": "error",
	}}
}

test_success_implicit_boolean_assignment if {
	r := rule.report with input as ast.policy(`allow { true }`)
	r == set()
}

test_fail_unification_in_default_assignment if {
	r := rule.report with input as ast.policy(`default x = false`)
	r == {{
		"category": "style",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "style"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "default x = false"},
		"level": "error",
	}}
}

test_success_assignment_in_default_assignment if {
	r := rule.report with input as ast.policy(`default x := false`)
	r == set()
}

test_fail_unification_in_object_rule_assignment if {
	r := rule.report with input as ast.policy(`x["a"] = 1`)
	r == {{
		"category": "style",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "style"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": `x["a"] = 1`},
		"level": "error",
	}}
}

test_success_assignment_in_object_rule_assignment if {
	r := rule.report with input as ast.policy(`x["a"] := 1`)
	r == set()
}

test_fail_unification_in_function_assignment if {
	r := rule.report with input as ast.policy(`foo(bar) = "baz"`)
	r == {{
		"category": "style",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "style"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": `foo(bar) = "baz"`},
		"level": "error",
	}}
}

test_success_implicit_boolean_assignment_function if {
	r := rule.report with input as ast.policy(`f(x) { 1 == 1 }`)
	r == set()
}

test_success_assignment_operator_function if {
	r := rule.report with input as ast.policy(`f(x) := true { 1 == 1 }`)
	r == set()
}

test_success_partial_rule if {
	r := rule.report with input as ast.policy(`partial["works"] { 1 == 1 }`)
	r == set()
}
