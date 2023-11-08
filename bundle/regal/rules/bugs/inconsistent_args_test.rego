package regal.rules.bugs["inconsistent-args_test"]

import future.keywords.if
import future.keywords.in

import data.regal.ast

import data.regal.rules.bugs["inconsistent-args"] as rule

test_fail_inconsistent_args if {
	module := ast.with_future_keywords(`
	foo(a, b) if a == b
	foo(b, a) if b > a

	bar(b, a) if b > a
	`)
	r := rule.report with input as module
	r == expected_with_location({"col": 2, "file": "policy.rego", "row": 10, "text": "\tfoo(b, a) if b > a"})
}

test_fail_nested_inconsistent_args if {
	module := ast.with_future_keywords(`
	a.b.foo(a, b) if a == b
	a.b.foo(b, a) if b > a
	`)
	r := rule.report with input as module
	r == expected_with_location({"col": 2, "file": "policy.rego", "row": 10, "text": "\ta.b.foo(b, a) if b > a"})
}

test_success_not_inconsistent_args if {
	module := ast.with_future_keywords(`
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
	module := ast.with_future_keywords(`
	foo(a, b) if a == b
	foo(_, b) if b.foo

	qux(c, a) if c == a
	`)
	r := rule.report with input as module
	r == set()
}

test_success_using_pattern_matching if {
	module := ast.with_future_keywords(`
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
		"ref": "https://docs.styra.com/regal/rules/bugs/inconsistent-args",
	}],
	"title": "inconsistent-args",
}

expected_with_location(location) := {object.union(expected, {"location": location})} if is_object(location)

expected_with_location(location) := {object.union(expected, {"location": loc}) |
	some loc in location
} if {
	is_set(location)
}
