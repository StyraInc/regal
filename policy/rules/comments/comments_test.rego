package regal.rules.comments_test

import future.keywords.if

import data.regal
import data.regal.rules.comments

test_fail_todo_comment if {
	ast := regal.ast(`# TODO: do someting clever`)
	result := comments.violation with input as ast
	result == {{
		"description": "TODO comment",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-comments-001"}],
		"title": "STY-COMMENTS-001",
        "location": {"col": 2, "file": "policy.rego", "row": 8},
	}}
}

test_fail_fixme_comment if {
	ast := regal.ast(`# fixme: this is broken`)
	result := comments.violation with input as ast
	result == {{
		"description": "TODO comment",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-comments-001"}],
		"title": "STY-COMMENTS-001",
        "location": {"col": 2, "file": "policy.rego", "row": 8},
	}}
}

test_success_no_todo_comment if {
	ast := regal.ast(`# This code is great`)
	result := comments.violation with input as ast
	result == set()
}
