package regal.rules.testing["dubious-print-sprintf_test"]

import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config

import data.regal.rules.testing["dubious-print-sprintf"] as rule

test_fail_print_sprintf {
	test_policy := ast.policy(`
y {
print(sprintf("name is: %s domain is: %s", [input.name, input.domain]))
}`)
  
	r := rule.report with input as test_policy
	r == {{
		"category": "testing",
		"description": "dubious use of print and sprintf",
		"level": "error",
		"location": {"col": 7, "file": "policy.rego", "row": 5, "text": "print(sprintf(\"name is: %s domain is: %s\", [input.name, input.domain]))"},		
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/dubious-print-sprintf", "testing"),
		}],
		"title": "dubious-print-sprintf"
	}}
}

test_fail_bodies_print_sprintf {
	test_policy := ast.policy(`
y {
comprehension := [x |
x := input[_]
print(sprintf("x is: %s", [x]))
]
}`)
  
	r := rule.report with input as test_policy
	r == {{
		"category": "testing",
		"description": "dubious use of print and sprintf",
		"level": "error",
		"location": {"col": 7, "file": "policy.rego", "row": 7, "text": "print(sprintf(\"x is: %s\", [x]))"},		
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/dubious-print-sprintf", "testing"),
		}],
		"title": "dubious-print-sprintf"
	}}
}
