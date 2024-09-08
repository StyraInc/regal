package regal.lsp.completion_test

import rego.v1

import data.regal.lsp.completion

test_completion_entrypoint if {
	items := completion.items with completion.providers as {"test": {"items": {{"foo": "bar"}}}}

	items == {{"_regal": {"provider": "test"}, "foo": "bar"}}
}
