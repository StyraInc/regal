package regal.lsp.completion.providers.import_test

import rego.v1

import data.regal.lsp.completion.providers["import"] as provider

test_import_completion_empty_line if {
	policy := `package policy

import rego.v1

`

	regal_module := {"regal": {
		"file": {
			"name": "p.rego",
			"lines": split(policy, "\n"),
		},
		"context": {"location": {"row": 5, "col": 1}},
	}}
	items := provider.items with input as regal_module

	items == {{
		"label": "import",
		"detail": "import <path>",
		"kind": 14,
		"textEdit": {
			"newText": "import ",
			"range": {
				"start": {"character": 0, "line": 4},
				"end": {"character": 0, "line": 4},
			},
		},
	}}
}

test_import_completion_on_typing if {
	policy := `package policy

import rego.v1

imp`

	regal_module := {"regal": {
		"file": {
			"name": "p.rego",
			"lines": split(policy, "\n"),
		},
		"context": {"location": {"row": 5, "col": 3}},
	}}
	items := provider.items with input as regal_module

	items == {{
		"label": "import",
		"detail": "import <path>",
		"kind": 14,
		"textEdit": {
			"newText": "import ",
			"range": {
				"start": {"character": 0, "line": 4},
				"end": {"character": 3, "line": 4},
			},
		},
	}}
}
