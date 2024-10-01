package regal.rules.bugs["invalid-metadata-attribute_test"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.rules.bugs["invalid-metadata-attribute"] as rule

test_fail_invalid_attribute if {
	r := rule.report with input as ast.policy(`
# METADATA
# title: allow
# is_true: yes
allow := true
`)

	r == {{
		"category": "bugs",
		"description": "Invalid attribute in metadata annotation",
		"level": "error",
		"location": {
			"col": 1,
			"file": "policy.rego",
			"row": 6,
			"end": {
				"col": 15,
				"row": 6,
			},
			"text": "# is_true: yes",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/invalid-metadata-attribute", "bugs"),
		}],
		"title": "invalid-metadata-attribute",
	}}
}

test_success_valid_metadata if {
	r := rule.report with input as ast.policy(`
# METADATA
# title: valid
# description: also valid
allow := true
`)

	r == set()
}
