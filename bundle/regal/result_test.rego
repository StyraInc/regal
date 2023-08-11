package regal.result_test

import data.regal.config
import data.regal.result

test_no_related_resources_in_result_fail_on_custom_rule_unless_provided {
	chain := [
		{"path": ["custom", "regal", "rules", "categoy", "name", "report"]},
		{
			"annotations": {
				"scope": "package",
				"description": "This is a test",
			},
			"path": ["custom", "regal", "rules", "categoy", "name"],
		},
	]

	violation := result.fail(chain, {})

	violation == {
		"category": "categoy",
		"description": "This is a test",
		"level": "error",
		"title": "name",
	}
}

test_related_resources_in_result_fail_on_custom_rule_when_provided {
	chain := [
		{"path": ["custom", "regal", "rules", "categoy", "name", "report"]},
		{
			"annotations": {
				"scope": "package",
				"description": "This is a test",
				"related_resources": [{
					"description": "documentation",
					"ref": "https://example.com",
				}],
			},
			"path": ["custom", "regal", "rules", "categoy", "name"],
		},
	]

	violation := result.fail(chain, {})

	violation == {
		"category": "categoy",
		"description": "This is a test",
		"level": "error",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://example.com",
		}],
		"title": "name",
	}
}

test_related_resources_generated_by_result_fail_for_builtin_rule {
	chain := [
		{"path": ["regal", "rules", "category", "name", "report"]},
		{
			"annotations": {
				"scope": "package",
				"description": "This is a test",
			},
			"path": ["regal", "rules", "category", "name"],
		},
	]

	violation := result.fail(chain, {})

	violation == {
		"category": "category",
		"description": "This is a test",
		"level": "error",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/name", "category"),
		}],
		"title": "name",
	}
}
