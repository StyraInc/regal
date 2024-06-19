package regal.lsp.completion.providers.locals_test

import rego.v1

import data.regal.lsp.completion.providers.locals as provider
import data.regal.lsp.completion.providers.utils_test as utils

test_no_locals_in_completion_items if {
	workspace := {"file:///p.rego": `package policy

import rego.v1

foo := 1

bar if {
	foo == 1
}
`}

	regal_module := {"regal": {
		"file": {
			"name": "p.rego",
			"lines": split(workspace["file:///p.rego"], "\n"),
		},
		"context": {"location": {
			"row": 8,
			"col": 9,
		}},
	}}
	items := provider.items with input as regal_module with data.workspace.parsed as utils.parsed_modules(workspace)

	count(items) == 0
}

test_locals_in_completion_items if {
	workspace := {"file:///p.rego": `package policy

import rego.v1

foo := 1

function(bar) if {
	baz := 1
	qux := b
}
`}

	regal_module := {"regal": {
		"file": {
			"name": "p.rego",
			"uri": "file:///p.rego",
			"lines": split(workspace["file:///p.rego"], "\n"),
		},
		"context": {"location": {
			"row": 9,
			"col": 10,
		}},
	}}

	items := provider.items with input as regal_module with data.workspace.parsed as utils.parsed_modules(workspace)

	count(items) == 2
	utils.expect_item(items, "bar", {"end": {"character": 9, "line": 8}, "start": {"character": 8, "line": 8}})
	utils.expect_item(items, "baz", {"end": {"character": 9, "line": 8}, "start": {"character": 8, "line": 8}})
}

test_locals_in_completion_items_function_call if {
	workspace := {"file:///p.rego": `package policy

import rego.v1

foo := 1

function(bar) if {
	baz := 1
	qux := other_function(b)
}
`}

	regal_module := {"regal": {
		"file": {
			"name": "p.rego",
			"uri": "file:///p.rego",
			"lines": split(workspace["file:///p.rego"], "\n"),
		},
		"context": {"location": {
			"row": 9,
			"col": 25,
		}},
	}}

	items := provider.items with input as regal_module with data.workspace.parsed as utils.parsed_modules(workspace)

	count(items) == 2
	utils.expect_item(items, "bar", {"end": {"character": 24, "line": 8}, "start": {"character": 23, "line": 8}})
	utils.expect_item(items, "baz", {"end": {"character": 24, "line": 8}, "start": {"character": 23, "line": 8}})
}

test_locals_in_completion_items_rule_head_assignment if {
	workspace := {"file:///p.rego": `package policy

import rego.v1

function(bar) := f if {
	foo := 1
}
`}

	regal_module := {"regal": {
		"file": {
			"name": "p.rego",
			"uri": "file:///p.rego",
			"lines": split(workspace["file:///p.rego"], "\n"),
		},
		"context": {"location": {
			"row": 5,
			"col": 19,
		}},
	}}
	items := provider.items with input as regal_module with data.workspace.parsed as utils.parsed_modules(workspace)

	count(items) == 1
	utils.expect_item(items, "foo", {"end": {"character": 18, "line": 4}, "start": {"character": 17, "line": 4}})
}
