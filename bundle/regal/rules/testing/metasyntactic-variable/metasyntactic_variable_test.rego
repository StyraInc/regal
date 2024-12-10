package regal.rules.testing["metasyntactic-variable_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.testing["metasyntactic-variable"] as rule

test_fail_rule_named_foo if {
	module := ast.policy("foo := true")

	r := rule.report with input as module
	r == {expected_with_location({
		"col": 1,
		"file": "policy.rego",
		"row": 3,
		"text": "foo := true",
		"end": {
			"col": 4,
			"row": 3,
		},
	})}
}

test_fail_metasyntactic_vars if {
	module := ast.policy(`allow if {
		fooBar := true
		input[baz]
	}`)

	r := rule.report with input as module
	r == {
		expected_with_location({
			"col": 3,
			"file": "policy.rego",
			"row": 4,
			"text": "\t\tfooBar := true",
			"end": {
				"col": 9,
				"row": 4,
			},
		}),
		expected_with_location({
			"col": 9,
			"file": "policy.rego",
			"row": 5,
			"text": "\t\tinput[baz]",
			"end": {
				"col": 12,
				"row": 5,
			},
		}),
	}
}

test_fail_metasyntactic_vars_ref_head_strings if {
	module := ast.policy(`foo.a.BAR.b.C.baz := true`)

	r := rule.report with input as module
	r == {
		expected_with_location({
			"col": 1,
			"file": "policy.rego",
			"row": 3,
			"text": "foo.a.BAR.b.C.baz := true",
			"end": {"col": 4, "row": 3},
		}),
		expected_with_location({
			"col": 7,
			"file": "policy.rego",
			"row": 3,
			"text": "foo.a.BAR.b.C.baz := true",
			"end": {"col": 10, "row": 3},
		}),
		expected_with_location({
			"col": 15,
			"file": "policy.rego",
			"row": 3,
			"text": "foo.a.BAR.b.C.baz := true",
			"end": {"col": 18, "row": 3},
		}),
	}
}

expected := {
	"category": "testing",
	"description": "Metasyntactic variable name",
	"level": "error",
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/metasyntactic-variable", "testing"),
	}],
	"title": "metasyntactic-variable",
}

# regal ignore:external-reference
expected_with_location(location) := object.union(expected, {"location": location})
