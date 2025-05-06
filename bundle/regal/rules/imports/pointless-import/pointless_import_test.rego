package regal.rules.imports["pointless-import_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.imports["pointless-import"] as rule

test_fail_pointless_import_of_same_package if {
	r := rule.report with input as ast.policy("import data.policy")

	r == {{
		"category": "imports",
		"description": "Importing own package is pointless",
		"level": "error",
		"location": {
			"col": 8,
			"end": {
				"col": 19,
				"row": 3,
			},
			"file": "policy.rego",
			"row": 3,
			"text": "import data.policy",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/pointless-import", "imports"),
		}],
		"title": "pointless-import",
	}}
}

test_fail_pointless_import_of_rule_in_same_package if {
	r := rule.report with input as ast.policy("import data.policy")

	r == {{
		"category": "imports",
		"description": "Importing own package is pointless",
		"level": "error",
		"location": {
			"col": 8,
			"end": {
				"col": 19,
				"row": 3,
			},
			"file": "policy.rego",
			"row": 3,
			"text": "import data.policy",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/pointless-import", "imports"),
		}],
		"title": "pointless-import",
	}}
}

test_success_somewhat_pointless_import_but_longer_ref_not_flagged if {
	r := rule.report with input as ast.policy("import data.policy.a.b.c.d.e")

	r == set()
}
