package regal.lsp.completion.providers.regov1_test

import data.regal.lsp.completion.providers.test_utils as util

import data.regal.lsp.completion.providers.regov1 as provider

test_regov1_completion_on_typing if {
	policy := `package policy

import r`
	items := provider.items with input as util.input_with_location_and_version(policy, {"row": 3, "col": 9}, 0)
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
	items := provider.items with input as util.input_with_location_and_version(policy, {"row": 3, "col": 8}, 0)
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
	items := provider.items with input as util.input_with_location_and_version(policy, {"row": 5, "col": 9}, 0)
	items == set()
}

test_no_regov1_completion_if_v1_file if {
	policy := `package policy

import r`
	items := provider.items with input as util.input_with_location_and_version(
		policy,
		{"row": 3, "col": 9},
		3, # RegoV1
	)
	items == set()
}
