package regal.rules.bugs["inconsistent-args_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["inconsistent-args"] as rule

test_fail_inconsistent_args if {
	module := ast.with_rego_v1(`
	foo(a, b) if a == b
	foo(b, a) if b > a

	bar(b, a) if b > a
	`)
	r := rule.report with input as module
	r == expected_with_location({
		"col": 6,
		"file": "policy.rego",
		"row": 7,
		"text": "\tfoo(b, a) if b > a",
		"end": {"col": 10, "row": 7},
	})
}

test_fail_nested_inconsistent_args if {
	module := ast.with_rego_v1(`
	a.b.foo(a, b) if a == b
	a.b.foo(b, a) if b > a
	`)
	r := rule.report with input as module
	r == expected_with_location({
		"col": 10,
		"file": "policy.rego",
		"row": 7,
		"text": "\ta.b.foo(b, a) if b > a",
		"end": {"col": 14, "row": 7},
	})
}

test_success_not_inconsistent_args if {
	module := ast.with_rego_v1(`
	foo(a, b) if a == b
	foo(a, b) if a > b

	bar(b, a) if b > a
	bar(b, a) if b == a

	qux(c, a) if c == a
	`)
	r := rule.report with input as module
	r == set()
}

test_success_using_wildcard if {
	module := ast.with_rego_v1(`
	foo(a, b) if a == b
	foo(_, b) if b.foo

	qux(c, a) if c == a
	`)
	r := rule.report with input as module
	r == set()
}

test_success_using_pattern_matching if {
	module := ast.with_rego_v1(`
	foo(a, b) if a == b
	foo(a, "foo") if a.foo

	qux(c, a) if c == a
	`)
	r := rule.report with input as module
	r == set()
}

expected := {
	"category": "bugs",
	"description": "Inconsistently named function arguments",
	"level": "error",
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/inconsistent-args", "bugs"),
	}],
	"title": "inconsistent-args",
}

# regal ignore:external-reference
expected_with_location(location) := {object.union(expected, {"location": location})} if is_object(location)

# regal ignore:external-reference
expected_with_location(location) := {object.union(expected, {"location": loc}) |
	some loc in location
} if {
	is_set(location)
}
