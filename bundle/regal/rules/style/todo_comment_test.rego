package regal.rules.style["todo-comment_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style["todo-comment"] as rule

test_fail_todo_comment if {
	r := rule.report with input as ast.policy(`# TODO: do someting clever`)
	r == {{
		"category": "style",
		"description": "Avoid TODO comments",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/todo-comment", "style"),
		}],
		"title": "todo-comment",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": `# TODO: do someting clever`},
		"level": "error",
	}}
}

test_fail_fixme_comment if {
	r := rule.report with input as ast.policy(`# fixme: this is broken`)
	r == {{
		"category": "style",
		"description": "Avoid TODO comments",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/todo-comment", "style"),
		}],
		"title": "todo-comment",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": `# fixme: this is broken`},
		"level": "error",
	}}
}

test_success_no_todo_comment if {
	r := rule.report with input as ast.policy(`# This code is great`)
	r == set()
}
