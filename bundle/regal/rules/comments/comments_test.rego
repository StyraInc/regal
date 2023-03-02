package regal.rules.comments_test

import future.keywords.if

import data.regal
import data.regal.rules.comments

test_fail_todo_comment if {
	report(`# TODO: do someting clever`) == {{
		"category": "comments",
		"description": "Avoid TODO comments",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/todo-comment"
		}],
		"title": "todo-comment",
        "location": {"col": 2, "file": "policy.rego", "row": 8},
	}}
}

test_fail_fixme_comment if {
	report(`# fixme: this is broken`) == {{
		"category": "comments",
		"description": "Avoid TODO comments",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/todo-comment"
		}],
		"title": "todo-comment",
        "location": {"col": 2, "file": "policy.rego", "row": 8},
	}}
}

test_success_no_todo_comment if {
	report(`# This code is great`) == set()
}

report(snippet) := report {
	report := comments.report with input as regal.ast(snippet) with regal.rule_config as {"enabled": true}
}
