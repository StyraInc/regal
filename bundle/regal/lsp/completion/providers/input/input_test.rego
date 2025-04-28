package regal.lsp.completion.providers.input_test

import data.regal.lsp.completion.providers.input as provider
import data.regal.lsp.completion.providers.test_utils as util

test_input_completion_on_typing if {
	policy := `package policy

allow if {
	i
}`
	items := provider.items with input as util.input_with_location(policy, {"row": 4, "col": 3})

	items == {{
		"detail": "input document",
		"documentation": {
			"kind": "markdown",
			"value": provider._doc,
		},
		"kind": 14,
		"label": "input",
		"textEdit": {
			"newText": "input",
			"range": {
				"end": {
					"character": 2,
					"line": 3,
				},
				"start": {
					"character": 1,
					"line": 3,
				},
			},
		},
	}}
}
