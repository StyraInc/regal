package regal.rules.imports["redundant-future-keywords_test"]

import data.regal.ast
import data.regal.config
import data.regal.rules.imports["redundant-future-keywords"] as rule

test_fail_future_keywords_every if {
	r := rule.report with input as ast.policy(`import future.keywords.every`)

	r == {{
		"category": "imports",
		"description": "Redundant import of future keywords",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/imports/redundant-future-keywords",
		}],
		"title": "redundant-future-keywords",
		"location": {
			"col": 8,
			"file": "policy.rego",
			"row": 3,
			"end": {
				"col": 14,
				"row": 3,
			},
			"text": `import future.keywords.every`,
		},
		"level": "error",
	}}
}

test_fail_future_keywords_in if {
	r := rule.report with input as ast.policy(`import future.keywords.in`)

	r == {{
		"category": "imports",
		"description": "Redundant import of future keywords",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/imports/redundant-future-keywords",
		}],
		"title": "redundant-future-keywords",
		"location": {
			"col": 8,
			"file": "policy.rego",
			"row": 3,
			"end": {
				"col": 14,
				"row": 3,
			},
			"text": `import future.keywords.in`,
		},
		"level": "error",
	}}
}

test_fail_future_keywords_if if {
	r := rule.report with input as ast.policy(`import future.keywords.if`)

	r == {{
		"category": "imports",
		"description": "Redundant import of future keywords",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/imports/redundant-future-keywords",
		}],
		"title": "redundant-future-keywords",
		"location": {
			"col": 8,
			"file": "policy.rego",
			"row": 3,
			"end": {
				"col": 14,
				"row": 3,
			},
			"text": `import future.keywords.if`,
		},
		"level": "error",
	}}
}

test_fail_future_keywords_contains if {
	r := rule.report with input as ast.policy(`import future.keywords.contains`)

	r == {{
		"category": "imports",
		"description": "Redundant import of future keywords",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/imports/redundant-future-keywords",
		}],
		"title": "redundant-future-keywords",
		"location": {
			"col": 8,
			"file": "policy.rego",
			"row": 3,
			"end": {
				"col": 14,
				"row": 3,
			},
			"text": `import future.keywords.contains`,
		},
		"level": "error",
	}}
}

test_success_future_keywords_not if {
	r := rule.report with input as ast.policy(`import future.keywords.not`)

	r == set()
}

test_success_future_keywords_and if {
	r := rule.report with input as ast.policy(`import future.keywords.and`)

	r == set()
}

test_success_future_keywords_or if {
	r := rule.report with input as ast.policy(`import future.keywords.or`)

	r == set()
}

test_success_older_capabilities_with_every if {
	r := rule.report with input as ast.policy(`import future.keywords.every`)
		with config.capabilities as {"future_keywords": ["contains", "every", "if", "in"]}

	r == set()
}
