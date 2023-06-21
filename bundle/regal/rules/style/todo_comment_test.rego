package regal.rules.style_test

import future.keywords.if

import data.regal.config
import data.regal.rules.style.common_test.report

test_fail_todo_comment if {
	report(`# TODO: do someting clever`) == {{
		"category": "style",
		"description": "Avoid TODO comments",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/todo-comment", "style"),
		}],
		"title": "todo-comment",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": `# TODO: do someting clever`},
		"level": "error",
	}}
}

test_fail_fixme_comment if {
	report(`# fixme: this is broken`) == {{
		"category": "style",
		"description": "Avoid TODO comments",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/todo-comment", "style"),
		}],
		"title": "todo-comment",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": `# fixme: this is broken`},
		"level": "error",
	}}
}

test_success_no_todo_comment if {
	report(`# This code is great`) == set()
}
