package regal.lsp.completion.providers.utils_test

import rego.v1

parsed_modules(workspace) := {file_uri: parsed_module |
	some file_uri, contents in workspace
	parsed_module := regal.parse_module(file_uri, contents)
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

input_module_with_location(module, policy, location) := object.union(module, {"regal": {
	"file": {
		"name": "p.rego",
		"lines": split(policy, "\n"),
	},
	"context": {"location": location},
}})

input_with_location(policy, location) := {"regal": {
	"file": {
		"name": "p.rego",
		"lines": split(policy, "\n"),
	},
	"context": {"location": location},
}}
