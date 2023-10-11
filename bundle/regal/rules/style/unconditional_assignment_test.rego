package regal.rules.style["unconditional-assignment_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style["unconditional-assignment"] as rule

test_fail_unconditional_assignment_in_body if {
	r := rule.report with input as ast.policy(`x := y {
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
		"location": {"col": 3, "file": "policy.rego", "row": 4, "text": "\t\ty := 1"},
		"level": "error",
	}}
}

test_fail_unconditional_eq_in_body if {
	r := rule.report with input as ast.policy(`x = y {
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
		"location": {"col": 3, "file": "policy.rego", "row": 4, "text": "\t\ty = 1"},
		"level": "error",
	}}
}

test_success_conditional_assignment_in_body if {
	r := rule.report with input as ast.policy(`x := y { input.foo == "bar"; y := 1 }`)
	r == set()
}

test_success_unconditional_assignment_but_with_in_body if {
	r := rule.report with input as ast.policy(`x := y { y := 5 with input as 1 }`)
	r == set()
}

test_success_unconditional_assignment_but_else if {
	r := rule.report with input as ast.policy(`msg := x {
    	x := input.foo
    } else := input.bar`)
	r == set()
}

test_fail_unconditional_multi_value_assignment if {
	r := rule.report with input as ast.with_future_keywords(`x contains y if {
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

test_fail_unconditional_map_assignment if {
	r := rule.report with input as ast.with_future_keywords(`x["y"] := y if {
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
