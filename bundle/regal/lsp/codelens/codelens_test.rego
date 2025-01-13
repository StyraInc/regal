package regal.lsp.codelens_test

import data.regal.lsp.codelens

# regal ignore:rule-length
test_code_lenses_for_module if {
	module := regal.parse_module("policy.rego", `
	package foo

	import rego.v1

	rule1 := 1

	rule2 if 1 + rule1 == 2
	`)
	lenses := codelens.lenses with input as module

	lenses == [
		{
			"command": {
				"arguments": ["policy.rego", "data.foo", 2],
				"command": "regal.eval",
				"title": "Evaluate",
			},
			"range": {"end": {"character": 8, "line": 1}, "start": {"character": 1, "line": 1}},
		},
		{
			"command": {
				"arguments": ["policy.rego", "data.foo.rule1", 6],
				"command": "regal.eval",
				"title": "Evaluate",
			},
			"range": {"end": {"character": 11, "line": 5}, "start": {"character": 1, "line": 5}},
		},
		{
			"command": {
				"arguments": ["policy.rego", "data.foo.rule2", 8],
				"command": "regal.eval", "title": "Evaluate",
			},
			"range": {"end": {"character": 24, "line": 7}, "start": {"character": 1, "line": 7}},
		},
		{
			"command": {
				"arguments": ["policy.rego", "data.foo", 2],
				"command": "regal.debug",
				"title": "Debug",
			},
			"range": {"end": {"character": 8, "line": 1}, "start": {"character": 1, "line": 1}},
		},
		{
			"command": {
				"arguments": ["policy.rego", "data.foo.rule2", 8],
				"command": "regal.debug",
				"title": "Debug",
			},
			"range": {"end": {"character": 24, "line": 7}, "start": {"character": 1, "line": 7}},
		},
	]
}
