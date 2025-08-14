package regal.lsp.documenthighlight_test

import data.regal.lsp.documenthighlight

test_metadata_header_highlight if {
	file_content := `package p

# METADATA
# description: A test rule
# scope: document
# title: Test Rule
allow if true`

	items := documenthighlight.items with input as {
		"params": {
			"textDocument": {"uri": "file://p.rego"},
			"position": {"line": 2, "character": 5},
		},
		"regal": {"file": {"lines": split(file_content, "\n")}},
	}
		with data.workspace.parsed["file://p.rego"] as regal.parse_module("p.rego", file_content)

	items == {
		# the METADATA itself
		{
			"range": {
				"start": {"line": 2, "character": 2},
				"end": {"line": 2, "character": 10},
			},
			"kind": 1,
		},
		# description
		{
			"range": {
				"start": {"line": 3, "character": 2},
				"end": {"line": 3, "character": 13},
			},
			"kind": 1,
		},
		# scope
		{
			"range": {
				"start": {"line": 4, "character": 2},
				"end": {"line": 4, "character": 7},
			},
			"kind": 1,
		},
		# title
		{
			"range": {
				"start": {"line": 5, "character": 2},
				"end": {"line": 5, "character": 7},
			},
			"kind": 1,
		},
	}
}

test_individual_metadata_attribute_highlight if {
	file_content := `package p

# METADATA
# description: A test rule
# scope: document
# title: Test Rule
allow if true`

	items := documenthighlight.items with input as {
		"params": {
			"textDocument": {"uri": "file://p.rego"},
			# over the description attribute
			"position": {"line": 3, "character": 5},
		},
		"regal": {"file": {"lines": split(file_content, "\n")}},
	}
		with data.workspace.parsed["file://p.rego"] as regal.parse_module("p.rego", file_content)

	# should highlight only the description attribute clicked
	items == {{
		"range": {
			"start": {"line": 3, "character": 2},
			"end": {"line": 3, "character": 13},
		},
		"kind": 1,
	}}
}
