package regal.rules.testing["dubious-print-sprintf_test"]

import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config

import data.regal.rules.testing["dubious-print-sprintf"] as rule

test_print_sprintf_not_allowed {
	test_policy := ast.policy(`
	y {
		print(sprintf("name is: %s domain is: %s", [input.name, input.domain]))
	}`)
  
	r := rule.report with input as test_policy

	r == {{
		"category": "testing",
		"description": "dubious print sprintf",
		"level": "error",
		"location": {"col": 9, "file": "policy.rego", "row": 5, "text": "\t\tprint(sprintf(\"name is: %s domain is: %s\", [input.name, input.domain]))"},		
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/dubious-print-sprintf", "testing"),
		}],
		"title": "dubious-print-sprintf"
	}}
}
