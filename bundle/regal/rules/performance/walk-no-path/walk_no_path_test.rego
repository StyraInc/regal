package regal.rules.performance["walk-no-path_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.performance["walk-no-path"] as rule

test_fail_walk_can_have_path_omitted if {
	module := ast.with_rego_v1(`
	allow if {
		walk(input, [path, value])
		value == 1
	}
	`)
	r := rule.report with input as module

	r == {with_location({
		"col": 3,
		"end": {
			"col": 7,
			"row": 7,
		},
		"file": "policy.rego",
		"row": 7,
		"text": "\t\twalk(input, [path, value])",
	})}
}

test_success_path_is_wildcard if {
	module := ast.with_rego_v1(`
	allow if {
		walk(input, [_, value])
		value == 1
	}
	`)
	r := rule.report with input as module

	r == set()
}

test_success_path_in_head if {
	module := ast.with_rego_v1(`
	rule[path] := "foo" if {
		walk(input, [path, value])
		value == 1
	}
	`)
	r := rule.report with input as module

	r == set()
}

test_success_path_in_call if {
	module := ast.with_rego_v1(`
	allow if {
		walk(input, [path, value])
		path == 1
	}
	`)
	r := rule.report with input as module

	r == set()
}

test_success_path_in_ref if {
	module := ast.with_rego_v1(`
	allow if {
		walk(input, [path, value])
		input[path] == 1
	}
	`)
	r := rule.report with input as module

	r == set()
}

test_success_path_in_ref_in_ref if {
	module := ast.with_rego_v1(`
	allow if {
		walk(input, [path, value])
		input[path[0]] == 1
	}
	`)
	r := rule.report with input as module

	r == set()
}

with_location(location) := {
	"category": "performance",
	"description": "Call to `walk` can be optimized",
	"level": "error",
	"location": location,
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/walk-no-path", "performance"),
	}],
	"title": "walk-no-path",
}
