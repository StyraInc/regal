package regal.lsp.completion.providers.location_test

import rego.v1

import data.regal.ast

import data.regal.lsp.completion.location

test_find_rule_from_location if {
	module := regal.parse_module("p.rego", `package p

import rego.v1

rule1 if {
	x := 1
}

rule2 if {
	y := 2
}

rule3 if {
	z := 3
}
`)
	not location.find_rule(module.rules, {"row": 2, "col": 6})

	r1 := location.find_rule(module.rules, {"row": 5, "col": 6})
	ast.ref_to_string(r1.head.ref) == "rule1"

	r2 := location.find_rule(module.rules, {"row": 9, "col": 11})
	ast.ref_to_string(r2.head.ref) == "rule2"

	r3 := location.find_rule(module.rules, {"row": 15, "col": 0})
	ast.ref_to_string(r3.head.ref) == "rule3"
}

test_find_locals_at_location if {
	module := regal.parse_module("p.rego", `package p

import rego.v1

rule if {
	x := 1
}

function(a, b) if {
	c := 3
}

another if {
	some x, y in collection
	z := x + y
}
`)

	location.find_locals(module.rules, {"row": 6, "col": 1}) with input as module == set()
	location.find_locals(module.rules, {"row": 6, "col": 10}) with input as module == {"x"}
	location.find_locals(module.rules, {"row": 10, "col": 1}) with input as module == {"a", "b"}
	location.find_locals(module.rules, {"row": 10, "col": 6}) with input as module == {"a", "b", "c"}
	location.find_locals(module.rules, {"row": 15, "col": 1}) with input as module == {"x", "y"}
	location.find_locals(module.rules, {"row": 16, "col": 1}) with input as module == {"x", "y", "z"}
}
