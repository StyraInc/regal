package regal.lsp.completion.providers.rulerefs_test

import data.regal.lsp.completion.providers.rulerefs
import rego.v1

workspace := {
	"current_file.rego": `package foo

import data.imported_pkg
import data.imported_pkg_2

local_rule := true
`,
	"imported_file.rego": `package imported_pkg

another_rule := true

_internal_rule := true
`,
	"imported_file_2.rego": `package imported_pkg_2

another_rule_2 := true

_internal_rule_2 := true
`,
	"not_imported_pkg.rego": `package not_imported_pkg.foo.bar

yet_another_rule := false

_internal_rule := true
`,
}

parsed_modules[file_uri] := parsed_module if {
	some file_uri, contents in workspace
	parsed_module := regal.parse_module(file_uri, contents)
}

test_rule_refs_no_word if {
	current_file_contents := concat("", [workspace["current_file.rego"], `
another_local_rule := `])

	regal_module := {"regal": {
		"file": {
			"name": "current_file.rego",
			"uri": "current_file.rego", # would be file:// prefixed in server
			"lines": split(current_file_contents, "\n"),
		},
		"context": {"location": {
			"row": 8,
			"col": 21,
		}},
	}}

	items := rulerefs.items with input as regal_module with data.workspace.parsed as parsed_modules

	labels := [item.label | some item in items]

	expected_refs := [
		"local_rule",
		"imported_pkg.another_rule",
		"imported_pkg_2.another_rule_2",
		"data.not_imported_pkg.foo.bar", # partial generated from rule below
		"data.not_imported_pkg.foo.bar.yet_another_rule",
	]

	expected_refs == labels
}

test_rule_refs_partial_word if {
	current_file_contents := concat("", [workspace["current_file.rego"], `
another_local_rule := imp`])

	regal_module := {"regal": {
		"file": {
			"name": "current_file.rego",
			"uri": "current_file.rego", # would be file:// prefixed in server
			"lines": split(current_file_contents, "\n"),
		},
		"context": {"location": {
			"row": 8,
			"col": 21,
		}},
	}}

	items := rulerefs.items with input as regal_module with data.workspace.parsed as parsed_modules

	labels := [item.label | some item in items]

	expected_refs := [
		"imported_pkg.another_rule",
		"imported_pkg_2.another_rule_2",
	]

	expected_refs == labels
}

test_rule_refs_not_in_rule if {
	current_file_contents := concat("", [workspace["current_file.rego"], `

a`])

	regal_module := {"regal": {
		"file": {
			"name": "current_file.rego",
			"uri": "current_file.rego", # would be file:// prefixed in server
			"lines": split(current_file_contents, "\n"),
		},
		"context": {"location": {
			"row": 8,
			"col": 21,
		}},
	}}

	items := rulerefs.items with input as regal_module with data.workspace.parsed as parsed_modules

	count(items) == 0
}
