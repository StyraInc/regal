package regal.lsp.completion.kind_test

import rego.v1

import data.regal.lsp.completion.kind

test_kind_for_coverage if {
	kind_values := [value | some value in kind]
	sort(kind_values) == numbers.range(1, 25)
}
