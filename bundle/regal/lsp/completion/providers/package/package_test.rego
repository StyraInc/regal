package regal.lsp.completion.providers.package_test

import rego.v1

import data.regal.lsp.completion.providers["package"] as provider
import data.regal.lsp.completion.providers.test_utils as util

test_package_completion_on_typing if {
	policy := `p`
	items := provider.items with input as util.input_with_location(policy, {"row": 1, "col": 2})
	items == {{
		"detail": "package <package-name>",
		"kind": 14,
		"label": "package",
		"textEdit": {
			"newText": "package ",
			"range": {
				"end": {
					"character": 1,
					"line": 0,
				},
				"start": {
					"character": 0,
					"line": 0,
				},
			},
		},
	}}
}

test_package_completion_on_invoked if {
	policy := ``
	items := provider.items with input as util.input_with_location(policy, {"row": 1, "col": 1})
	items == {{
		"detail": "package <package-name>",
		"kind": 14,
		"label": "package",
		"textEdit": {
			"newText": "package ",
			"range": {
				"end": {
					"character": 0,
					"line": 0,
				},
				"start": {
					"character": 0,
					"line": 0,
				},
			},
		},
	}}
}

test_package_completion_not_suggested_if_already_present if {
	policy := `packae policy

	`
	items := provider.items with input as util.input_with_location(policy, {"row": 3, "col": 1})
	items == set()
}
