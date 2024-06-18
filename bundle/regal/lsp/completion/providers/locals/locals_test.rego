package regal.lsp.completion.providers.locals_test

import rego.v1

import data.regal.lsp.completion.providers.locals

test_no_locals_in_completion_items if {
	policy := `package policy

import rego.v1

foo := 1

bar if {
	foo == 1
}
`

	module := regal.parse_module("p.rego", policy)
	regal_module := object.union(module, {"regal": {
		"file": {
			"name": "p.rego",
			"lines": split(policy, "\n"),
		},
		"context": {"location": {
			"row": 8,
			"col": 9,
		}},
	}})
	items := locals.items with input as regal_module

	count(items) == 0
}

test_locals_in_completion_items if {
	policy := `package policy

import rego.v1

foo := 1

function(bar) if {
	baz := 1
	qux := b
}
`

	module := object.union(regal.parse_module("p.rego", policy), {"regal": {
		"file": {
			"name": "p.rego",
			"lines": split(policy, "\n"),
		},
		"context": {"location": {
			"row": 9,
			"col": 10,
		}},
	}})
	items := locals.items with input as module

	count(items) == 2
	expect_item(items, "bar", {"end": {"character": 9, "line": 8}, "start": {"character": 8, "line": 8}})
	expect_item(items, "baz", {"end": {"character": 9, "line": 8}, "start": {"character": 8, "line": 8}})
}

test_locals_in_completion_items_function_call if {
	policy := `package policy

import rego.v1

foo := 1

function(bar) if {
	baz := 1
	qux := other_function(b)
}
`
	module := object.union(regal.parse_module("p.rego", policy), {"regal": {
		"file": {
			"name": "p.rego",
			"lines": split(policy, "\n"),
		},
		"context": {"location": {
			"row": 9,
			"col": 25,
		}},
	}})
	items := locals.items with input as module

	count(items) == 2
	expect_item(items, "bar", {"end": {"character": 24, "line": 8}, "start": {"character": 23, "line": 8}})
	expect_item(items, "baz", {"end": {"character": 24, "line": 8}, "start": {"character": 23, "line": 8}})
}

test_locals_in_completion_items_rule_head_assignment if {
	policy := `package policy

import rego.v1

function(bar) := f if {
	foo := 1
}
`
	module := object.union(regal.parse_module("p.rego", policy), {"regal": {
		"file": {
			"name": "p.rego",
			"lines": split(policy, "\n"),
		},
		"context": {"location": {
			"row": 5,
			"col": 19,
		}},
	}})
	items := locals.items with input as module

	count(items) == 1
	expect_item(items, "foo", {"end": {"character": 18, "line": 4}, "start": {"character": 17, "line": 4}})
}

expect_item(items, label, range) if {
	expected := {"detail": "local variable", "kind": 6}

	item := object.union(expected, {
		"label": label,
		"textEdit": {
			"newText": label,
			"range": range,
		},
	})

	item in items
}
