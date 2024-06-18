package regal.lsp.completion.providers.import_test

import rego.v1

import data.regal.lsp.completion.providers["import"] as provider

test_import_completion_on_typing if {
	policy := `package policy

import rego.v1

`
	module := regal.parse_module("p.rego", policy)
	new_policy := concat("", [policy, "i"])
	items := provider.items with input as input_with_location(module, new_policy, {"row": 5, "col": 2})

	items == {{
		"label": "import",
		"detail": "import <path>",
		"kind": 14,
		"textEdit": {
			"newText": "import ",
			"range": {
				"start": {"character": 0, "line": 4},
				"end": {"character": 1, "line": 4},
			},
		},
	}}
}

input_with_location(module, policy, location) := object.union(module, {"regal": {
	"file": {
		"name": "p.rego",
		"lines": split(policy, "\n"),
	},
	"context": {"location": location},
}})
