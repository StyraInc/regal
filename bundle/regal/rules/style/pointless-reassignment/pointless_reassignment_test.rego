package regal.rules.style["pointless-reassignment_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.style["pointless-reassignment"] as rule

test_pointless_reassignment_in_rule_head if {
	module := ast.with_rego_v1(`
	foo := "foo"

	bar := foo
	`)

	r := rule.report with input as module
	r == {{
		"category": "style",
		"description": "Pointless reassignment of variable",
		"level": "error",
		"location": {
			"col": 2,
			"file": "policy.rego",
			"row": 8,
			"text": "\tbar := foo",
			"end": {
				"col": 12,
				"row": 8,
			},
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/pointless-reassignment", "style"),
		}],
		"title": "pointless-reassignment",
	}}
}

test_pointless_reassignment_in_rule_body if {
	module := ast.with_rego_v1(`
	rule if {
		foo := "foo"

		bar := foo
	}
	`)

	r := rule.report with input as module
	r == {{
		"category": "style",
		"description": "Pointless reassignment of variable",
		"level": "error",
		"location": {
			"col": 3,
			"file": "policy.rego",
			"row": 9,
			"end": {
				"col": 13,
				"row": 9,
			},
			"text": "\t\tbar := foo",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/pointless-reassignment", "style"),
		}],
		"title": "pointless-reassignment",
	}}
}

test_pointless_reassignment_in_rule_body_using_with if {
	module := ast.with_rego_v1(`
	foo := input

	rule if {
		bar := foo with input as "wow"

		bar == true
	}
	`)

	r := rule.report with input as module
	r == set()
}
