package regal.lsp.completion.providers.packagename_test

import rego.v1

import data.regal.lsp.completion.providers.packagename as provider

test_package_name_completion_on_typing if {
	policy := `package f`
	provider_input := {"regal": {
		"file": {
			"name": "/Users/joe/policy/foo/bar/baz/p.rego",
			"lines": split(policy, "\n"),
		},
		"context": {
			"path_separator": "/",
			"workspace_root": "/Users/joe/policy",
			"location": {
				"row": 1,
				"col": 10,
			},
		},
	}}
	items := provider.items with input as provider_input
	items == {{
		"detail": "suggested package name based on directory structure",
		"kind": 19,
		"label": "foo.bar.baz",
		"textEdit": {
			"newText": "foo.bar.baz\n\n",
			"range": {
				"end": {"character": 9, "line": 0},
				"start": {"character": 8, "line": 0},
			},
		},
	}}
}

# regal ignore:rule-length
test_package_name_completion_on_typing_multiple_suggestions if {
	policy := `package b`
	provider_input := {"regal": {
		"file": {
			"name": "/Users/joe/policy/foo/bar/baz/p.rego",
			"lines": split(policy, "\n"),
		},
		"context": {
			"path_separator": "/",
			"workspace_root": "/Users/joe/policy",
			"location": {
				"row": 1,
				"col": 10,
			},
		},
	}}
	items := provider.items with input as provider_input
	items == {
		{
			"detail": "suggested package name based on directory structure",
			"kind": 19,
			"label": "bar.baz",
			"textEdit": {
				"newText": "bar.baz\n\n",
				"range": {
					"end": {"character": 9, "line": 0},
					"start": {"character": 8, "line": 0},
				},
			},
		},
		{
			"detail": "suggested package name based on directory structure",
			"kind": 19,
			"label": "baz",
			"textEdit": {
				"newText": "baz\n\n",
				"range": {
					"end": {"character": 9, "line": 0},
					"start": {"character": 8, "line": 0},
				},
			},
		},
	}
}

# regal ignore:rule-length
test_package_name_completion_on_typing_multiple_suggestions_when_invoked if {
	policy := `package `
	provider_input := {"regal": {
		"file": {
			"name": "/Users/joe/policy/foo/bar/baz/p.rego",
			"lines": split(policy, "\n"),
		},
		"context": {
			"path_separator": "/",
			"workspace_root": "/Users/joe/policy",
			"location": {
				"row": 1,
				"col": 9,
			},
		},
	}}
	items := provider.items with input as provider_input
	items == {
		{
			"detail": "suggested package name based on directory structure",
			"kind": 19,
			"label": "foo.bar.baz",
			"textEdit": {
				"newText": "foo.bar.baz\n\n",
				"range": {
					"end": {"character": 8, "line": 0},
					"start": {"character": 8, "line": 0},
				},
			},
		},
		{
			"detail": "suggested package name based on directory structure",
			"kind": 19,
			"label": "bar.baz",
			"textEdit": {
				"newText": "bar.baz\n\n",
				"range": {
					"end": {"character": 8, "line": 0},
					"start": {"character": 8, "line": 0},
				},
			},
		},
		{
			"detail": "suggested package name based on directory structure",
			"kind": 19,
			"label": "baz",
			"textEdit": {
				"newText": "baz\n\n",
				"range": {
					"end": {"character": 8, "line": 0},
					"start": {"character": 8, "line": 0},
				},
			},
		},
	}
}

test_build_suggestions if {
	provider._suggestions("foo.bar.baz", {"text": "foo"}) == ["foo.bar.baz"]

	provider._suggestions("foo.bar.baz", {"text": "bar"}) == ["bar.baz"]

	provider._suggestions("foo.bar.baz", {"text": "ba"}) == ["bar.baz", "baz"]
}

test_build_suggestions_invoked if {
	provider._suggestions("foo.bar.baz", {"text": ""}) == [
		"foo.bar.baz",
		"bar.baz",
		"baz",
	]
}
