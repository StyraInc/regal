package regal.rules.style["no-whitespace-comment_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style["no-whitespace-comment"] as rule

test_no_leading_whitespace if {
	r := rule.report with input as ast.policy(`#foo`)
	r == {{
		"category": "style",
		"description": "Comment should start with whitespace",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/no-whitespace-comment", "style"),
		}],
		"title": "no-whitespace-comment",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "#foo"},
		"level": "error",
	}}
}

test_success_leading_whitespace if {
	r := rule.report with input as ast.policy(`# foo`)
	r == set()
}

test_success_lonely_hash if {
	r := rule.report with input as ast.policy(`#`)
	r == set()
}

test_success_comment_with_newline if {
	r := rule.report with input as ast.policy(`
	# foo
	#
	# bar`)
	r == set()
}
