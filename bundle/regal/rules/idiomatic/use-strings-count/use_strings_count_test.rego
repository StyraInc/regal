package regal.rules.idiomatic["use-strings-count_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.idiomatic["use-strings-count"] as rule

test_fail_can_use_strings_count if {
	module := ast.with_rego_v1(`x := count(indexof_n("foo", "o"))`)

	r := rule.report with input as module
	r == {{
		"category": "idiomatic",
		"description": "Use `strings.count` where possible",
		"level": "error",
		"location": {
			"col": 6,
			"file": "policy.rego",
			"row": 5,
			"text": `x := count(indexof_n("foo", "o"))`,
			"end": {"col": 34, "row": 5},
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-strings-count", "idiomatic"),
		}],
		"title": "use-strings-count",
	}}
}

test_has_notice_if_unmet_capability if {
	r := rule.notices with config.capabilities as {}
	r == {{
		"category": "idiomatic",
		"description": "Missing capability for built-in function `strings.count`",
		"level": "notice",
		"severity": "warning",
		"title": "use-strings-count",
	}}
}
