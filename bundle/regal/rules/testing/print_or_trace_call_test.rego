package regal.rules.testing["print-or-trace-call_test"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.rules.testing["print-or-trace-call"] as rule

test_fail_call_to_print_and_trace if {
	r := rule.report with input as ast.policy(`allow {
		print("foo")

		x := [i | i = 0; trace("bar")]
	}`)
	r == {
		expected_with_location({
			"col": 3,
			"file": "policy.rego",
			"row": 4,
			"text": "\t\tprint(\"foo\")",
		}),
		expected_with_location({
			"col": 20,
			"file": "policy.rego",
			"row": 6,
			"text": "\t\tx := [i | i = 0; trace(\"bar\")]",
		}),
	}
}

expected_with_location(location) := {
	"category": "testing",
	"description": "Call to print or trace function",
	"level": "error",
	"location": location,
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/print-or-trace-call", "testing"),
	}],
	"title": "print-or-trace-call",
}
