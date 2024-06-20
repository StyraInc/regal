package regal.lsp.completion.providers.commonrule_test

import rego.v1

import data.regal.lsp.completion.providers.commonrule as provider
import data.regal.lsp.completion.providers.utils_test as util

test_common_name_completion_on_invoked if {
	policy := `package policy

import rego.v1


`
	module := regal.parse_module("p.rego", policy)
	items := provider.items with input as util.input_module_with_location(module, policy, {"row": 5, "col": 2})

	expected_item(items, "allow")
	expected_item(items, "deny")
	expected_item(items, "authorized")
}

test_common_name_completion_on_typed if {
	policy := `package policy

import rego.v1


`
	module := regal.parse_module("p.rego", policy)
	new_policy := concat("", [policy, "d"])
	items := provider.items with input as util.input_module_with_location(module, new_policy, {"row": 5, "col": 2})

	expected_item(items, "deny")
}

expected_item(items, label) if {
	item := {
		"label": label,
		"detail": "common name",
		"documentation": {
			"kind": "markdown",
			"value": sprintf("%q is a common rule name", [label]),
		},
		"kind": 15,
		"textEdit": {
			"range": {
				"start": {
					"line": 4,
					"character": 0,
				},
				"end": {
					"line": 4,
					"character": 1,
				},
			},
			"newText": sprintf("%s ", [label]),
		},
	}

	item in items
}
