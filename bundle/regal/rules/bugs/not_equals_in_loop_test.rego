package regal.rules.bugs_test

import future.keywords.if

import data.regal.rules.bugs.common_test.report
import data.regal.rules.bugs.common_test.report_with_fk

test_fail_neq_in_loop if {
	r := report(`deny {
		"admin" != input.user.groups[_]
		input.user.groups[_] != "admin"
	}`)
	r == {
		{
			"category": "bugs",
			"description": "Use of != in loop",
			"level": "error",
			"location": {"col": 11, "file": "policy.rego", "row": 4, "text": "\t\t\"admin\" != input.user.groups[_]"},
			"related_resources": [{
				"description": "documentation",
				"ref": "https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/not-equals-in-loop.md",
			}],
			"title": "not-equals-in-loop",
		},
		{
			"category": "bugs",
			"description": "Use of != in loop",
			"level": "error",
			"location": {"col": 24, "file": "policy.rego", "row": 5, "text": "\t\tinput.user.groups[_] != \"admin\""},
			"related_resources": [{
				"description": "documentation",
				"ref": "https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/not-equals-in-loop.md",
			}],
			"title": "not-equals-in-loop",
		},
	}
}

test_fail_neq_in_loop_one_liner if {
	r := report_with_fk(`deny if "admin" != input.user.groups[_]`)
	r == {{
		"category": "bugs",
		"description": "Use of != in loop",
		"level": "error",
		"location": {"col": 17, "file": "policy.rego", "row": 8, "text": "deny if \"admin\" != input.user.groups[_]"},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/not-equals-in-loop.md",
		}],
		"title": "not-equals-in-loop",
	}}
}
