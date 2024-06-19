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
