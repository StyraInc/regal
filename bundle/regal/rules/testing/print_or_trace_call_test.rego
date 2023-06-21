package regal.rules.testing_test

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.rules.testing.common_test.report

test_fail_call_to_print_and_trace if {
	r := report(`allow {
		print("foo")

		x := [i | i = 0; trace("bar")]
	}`)
	r == {
		{
			"category": "testing",
			"description": "Call to print or trace function",
			"level": "error",
			"location": {"col": 3, "file": "policy.rego", "row": 9, "text": "\t\tprint(\"foo\")"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/print-or-trace-call", "testing"),
			}],
			"title": "print-or-trace-call",
		},
		{
			"category": "testing",
			"description": "Call to print or trace function",
			"level": "error",
			"location": {"col": 20, "file": "policy.rego", "row": 11, "text": "\t\tx := [i | i = 0; trace(\"bar\")]"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/print-or-trace-call", "testing"),
			}],
			"title": "print-or-trace-call",
		},
	}
}
