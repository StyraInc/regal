package regal.rules.style["messy-rule_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.style["messy-rule"] as rule

test_success_non_messy_definition if {
	module := ast.with_rego_v1(`
	foo if true

	foo if 5 == 1

	bar if false
	`)
	r := rule.report with input as module

	r == set()
}

test_fail_messy_definition if {
	module := ast.with_rego_v1(`
	foo if true

	bar if false

	foo if 5 == 1
	`)
	r := rule.report with input as module

	r == expected_with_location({
		"col": 2,
		"row": 10,
		"end": {
			"col": 15,
			"row": 10,
		},
		"text": "\tfoo if 5 == 1",
	})
}

test_fail_messy_default_definition if {
	module := ast.with_rego_v1(`
	default foo := true

	bar if false

	foo if 5 == 1
	`)
	r := rule.report with input as module

	r == expected_with_location({
		"col": 2,
		"row": 10,
		"end": {
			"col": 15,
			"row": 10,
		},
		"text": "\tfoo if 5 == 1",
	})
}

test_fail_messy_nested_rule_definition if {
	module := ast.with_rego_v1(`
	base.foo if true

	bar if false

	base.foo if 5 == 1
	`)
	r := rule.report with input as module

	r == expected_with_location({
		"col": 2,
		"row": 10,
		"end": {
			"col": 20,
			"row": 10,
		},
		"text": "\tbase.foo if 5 == 1",
	})
}

test_success_non_incremental_nested_rule_definition if {
	module := ast.with_rego_v1(`
	base.foo if true

	bar if false

	base.bar if 5 == 1
	`)
	r := rule.report with input as module

	r == set()
}

test_success_non_messy_ref_head_rules if {
	module := ast.with_rego_v1(`
	keywords[foo.bar] contains "foo"

	keywords[foo] contains "foo"

	keywords[foo.baz] contains "foo"
	`)

	r := rule.report with input as module
	r == set()
}

test_fail_messy_incremental_nested_variable_rule_definitiion if {
	module := ast.with_rego_v1(`
	base[x].foo := 5 if { x := 1 }

	bar if false

	base[x].foo := 1 if { x := 1 }
	`)
	r := rule.report with input as module

	r == expected_with_location({
		"col": 2,
		"row": 10,
		"end": {
			"col": 32,
			"row": 10,
		},
		"text": "\tbase[x].foo := 1 if { x := 1 }",
	})
}

expected := {
	"category": "style",
	"description": "Messy incremental rule",
	"level": "error",
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/messy-rule", "style"),
	}],
	"title": "messy-rule",
	"location": {"file": "policy.rego"},
}

# regal ignore:external-reference
expected_with_location(location) := {object.union(expected, {"location": location})} if is_object(location)
