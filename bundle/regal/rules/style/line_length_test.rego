package regal.rules.style_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style
import data.regal.rules.style.common_test.report

test_fail_line_too_long if {
	r := report(`allow {
foo == bar; bar == baz; [a, b, c, d, e, f] := [1, 2, 3, 4, 5, 6]; qux := [q | some q in input.nonsense]
	}`)
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
	report(`allow { "foo" == "bar" }`) == set()
}
