package regal.rules.style["line-length_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style["line-length"] as rule

test_fail_line_too_long if {
	r := rule.report with input as ast.with_future_keywords(`allow {
foo == bar; bar == baz; [a, b, c, d, e, f] := [1, 2, 3, 4, 5, 6]; qux := [q | some q in input.nonsense]
	}`)
		with config.for_rule as {"level": "error", "max-line-length": 80}
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
	r := rule.report with input as ast.policy(`allow { "foo" == "bar" }`)
		with config.for_rule as {"level": "error", "max-line-length": 80}
	r == set()
}
