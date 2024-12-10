package regal.ast_test

import data.regal.ast

test_exprs if {
	inp := regal.parse_module("foo.rego", `package example

import rego.v1

allow if input.x == 1

allow if {
	input.y == 2
	input.z == 3
}
`)

	result := ast.exprs with input as inp

	count(result) == 2 # rules

	count(result[0]) == 1
	result[0][0].terms[0].value[0].value == "equal"

	count(result[1]) == 2
	result[1][0].terms[0].value[0].value == "equal"
	result[1][1].terms[0].value[0].value == "equal"
}
