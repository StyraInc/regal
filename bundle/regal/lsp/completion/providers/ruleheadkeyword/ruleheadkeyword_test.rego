package regal.lsp.completion.providers.ruleheadkeyword_test

import data.regal.lsp.completion.providers.ruleheadkeyword as provider

test_keyword_completion_after_rule_name_no_prefix[label] if {
	items := provider.items with input as {"regal": {
		"file": {
			"name": "/ws/p.rego",
			"lines": split("package p\n\nrule ", "\n"),
		},
		"context": {
			"workspace_root": "/ws",
			"location": {
				"row": 3,
				"col": 6,
			},
		},
		"environment": {"path_separator": "/"},
	}}

	count(items) == 3

	some label, completion in provider.completions
	expected := object.union(completion, {"textEdit": {
		"newText": sprintf("%s ", [label]),
		"range": {
			"start": {"line": 2, "character": 5},
			"end": {"line": 2, "character": 5},
		},
	}})

	expected in items
}

test_keyword_completion_after_rule_name_i_prefix_suggests_only_if if {
	items := provider.items with input as {"regal": {
		"file": {
			"name": "/ws/p.rego",
			"lines": split("package p\n\nrule i", "\n"),
		},
		"context": {
			"workspace_root": "/ws",
			"location": {
				"row": 3,
				"col": 7,
			},
		},
		"environment": {"path_separator": "/"},
	}}

	items == {object.union(provider.completions["if"], {"textEdit": {
		"newText": "if ",
		"range": {
			"start": {"line": 2, "character": 5},
			"end": {"line": 2, "character": 6},
		},
	}})}
}

test_completion_after_contains_only_has_if if {
	items := provider.items with input as {"regal": {
		"file": {
			"name": "/ws/p.rego",
			"lines": split("package p\n\nrule contains 100 ", "\n"),
		},
		"context": {
			"workspace_root": "/ws",
			"location": {
				"row": 3,
				"col": 19,
			},
		},
		"environment": {"path_separator": "/"},
	}}

	expected := {{
		"kind": 14,
		"label": "if",
		"labelDetails": {"description": "add conditions for rule to evaluate"},
		"textEdit": {
			"newText": "if ",
			"range": {
				"end": {"character": 18, "line": 2},
				"start": {"character": 18, "line": 2},
			},
		},
	}}

	items == expected
}
