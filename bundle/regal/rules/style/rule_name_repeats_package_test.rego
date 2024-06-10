package regal.rules.style["rule-name-repeats-package_test"]

import rego.v1

import data.regal.config

import data.regal.rules.style["rule-name-repeats-package"] as rule

related_resources := [{
	"description": "documentation",
	"ref": config.docs.resolve_url("$baseUrl/$category/rule-name-repeats-package", "style"),
}]

base_result := {
	"title": "rule-name-repeats-package",
	"description": "Rule name repeats package",
	"category": "style", "level": "error",
	"related_resources": related_resources,
}

test_possible_offending_prefixes if {
	module := regal.parse_module("example.rego", `
    package policy.foo.bar

    allow := true
    `)

	r := rule.possible_offending_prefixes with input as module

	r == {
		"policy_foo_bar",
		"foo_bar",
		"bar",
		"policyFooBar",
		"fooBar",
	}
}

test_rule_empty_if_no_repetition if {
	module := regal.parse_module("example.rego", `
    package policy.foo.bar

    allow := true
    `)

	r := rule.report with input as module

	r == set()
}

test_rule_violation_if_repetition if {
	module := regal.parse_module("example.rego", `
    package policy.foo.bar

    bar := true
    `)

	r := rule.report with input as module

	r == {object.union(
		base_result,
		{"location": {"col": 5, "file": "example.rego", "row": 4, "text": "    bar := true"}},
	)}
}

test_rule_violation_if_repetition_of_more_than_one_path_component if {
	module := regal.parse_module("example.rego", `
    package policy.foo.bar.baz

    foo_bar_baz := true

    barBaz := 1
    `)

	r := rule.report with input as module

	r == {
		object.union(
			base_result,
			{"location": {"col": 5, "file": "example.rego", "row": 4, "text": "    foo_bar_baz := true"}},
		),
		object.union(
			base_result,
			{"location": {"col": 5, "file": "example.rego", "row": 6, "text": "    barBaz := 1"}},
		),
	}
}

test_rule_violation_if_repetition_multiple if {
	module := regal.parse_module("example.rego", `
    package policy.foo.bar

    bar := true
    barNumber := 3
    barString := "string"
    `)

	r := rule.report with input as module

	r == {
		object.union(base_result, {"location": {"col": 5, "file": "example.rego", "row": 4, "text": "    bar := true"}}),
		object.union(base_result, {"location": {"col": 5, "file": "example.rego", "row": 5, "text": "    barNumber := 3"}}),
		object.union(
			base_result,
			{"location": {"col": 5, "file": "example.rego", "row": 6, "text": "    barString := \"string\""}},
		),
	}
}

test_rule_violation_if_repetition_in_function if {
	module := regal.parse_module("example.rego", `
    package policy.foo.bar

    bar(_) := true
    `)

	r := rule.report with input as module

	r == {object.union(
		base_result,
		{"location": {"col": 5, "file": "example.rego", "row": 4, "text": "    bar(_) := true"}},
	)}
}

test_rule_violation_if_repetition_in_defaults if {
	module := regal.parse_module("example.rego", `
    package policy.foo.bar

    default bar(_) := true
    default barNumber := 3
    `)

	r := rule.report with input as module

	r == {
		object.union(
			base_result,
			{"location": {"col": 13, "file": "example.rego", "row": 4, "text": "    default bar(_) := true"}},
		),
		object.union(
			base_result,
			{"location": {"col": 13, "file": "example.rego", "row": 5, "text": "    default barNumber := 3"}},
		),
	}
}
