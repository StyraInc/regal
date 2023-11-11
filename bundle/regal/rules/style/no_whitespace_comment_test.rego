package regal.rules.style["no-whitespace-comment_test"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.rules.style["no-whitespace-comment"] as rule

test_fail_no_leading_whitespace if {
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

test_fail_no_leading_whitespace_multiple_hashes if {
	r := rule.report with input as ast.policy(`##foo`)
	r == {{
		"category": "style",
		"description": "Comment should start with whitespace",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/no-whitespace-comment", "style"),
		}],
		"title": "no-whitespace-comment",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "##foo"},
		"level": "error",
	}}
}

test_success_excepted_pattern if {
	r := rule.report with input as ast.policy(`#-- foo`) with config.for_rule as {"except-pattern": "^--"}
	r == set()
}

test_success_leading_whitespace if {
	r := rule.report with input as ast.policy(`# foo`)
	r == set()
}

test_success_leading_whitespace_double_hash if {
	r := rule.report with input as ast.policy(`## foo`)
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

test_success_multiple_hash_comment if {
	r := rule.report with input as ast.policy(`
	##########
	# Foobar #
	##########`)
	r == set()
}
