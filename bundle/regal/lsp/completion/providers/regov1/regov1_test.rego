package regal.lsp.completion.providers.regov1_test

import rego.v1

import data.regal.lsp.completion.providers.utils_test as util

import data.regal.lsp.completion.providers.regov1 as provider

test_regov1_completion_on_typing if {
	policy := `package policy

import r`
	items := provider.items with input as util.input_with_location(policy, {"row": 3, "col": 9})
	items == {{
		"label": "rego.v1",
		"kind": 9,
		"detail": "use rego.v1",
		"textEdit": {
			"range": {
				"start": {
					"line": 2,
					"character": 7,
				},
				"end": {
					"line": 2,
					"character": 8,
				},
			},
			"newText": "rego.v1\n\n",
		},
	}}
}

test_regov1_completion_on_invoked if {
	policy := `package policy

import `
	items := provider.items with input as util.input_with_location(policy, {"row": 3, "col": 8})
	items == {{
		"label": "rego.v1",
		"kind": 9,
		"detail": "use rego.v1",
		"textEdit": {
			"range": {
				"start": {
					"line": 2,
					"character": 7,
				},
				"end": {
					"line": 2,
					"character": 7,
				},
			},
			"newText": "rego.v1\n\n",
		},
	}}
}

test_no_regov1_completion_if_already_imported if {
	policy := `package policy

import rego.v1

import r`
	items := provider.items with input as util.input_with_location(policy, {"row": 5, "col": 9})
	items == set()
}
