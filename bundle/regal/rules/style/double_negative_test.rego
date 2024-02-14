package regal.rules.style["double-negative_test"]

import rego.v1

import data.regal.ast

import data.regal.rules.style["double-negative"] as rule

test_fail_double_negative if {
	r := rule.report with input as ast.policy(`
    import future.keywords.if

    not_fine := true

    fine if not not_fine
    `)
	r == {{
		"category": "style",
		"description": "Avoid double negatives",
		"location": {"col": 13, "file": "policy.rego", "row": 8, "text": "    fine if not not_fine"},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/community/double-negative",
		}],
		"title": "double-negative",
		"level": "error",
	}}
}
