package regal.lsp.completion.location_test

import rego.v1

import data.regal.ast

import data.regal.lsp.completion.location

# regal ignore:rule-length
test_find_rule_from_location if {
	policy := `package p

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
`
	lines := split(policy, "\n")

	module := regal.parse_module("p.rego", policy)

	not location.find_rule(module.rules, {"row": 2, "col": 6}) with input.regal.file.lines as lines

	r1 := location.find_rule(module.rules, {"row": 5, "col": 6}) with input.regal.file.lines as lines

	ast.ref_to_string(r1.head.ref) == "rule1"

	r2 := location.find_rule(module.rules, {"row": 9, "col": 11}) with input.regal.file.lines as lines
	ast.ref_to_string(r2.head.ref) == "rule2"

	r3 := location.find_rule(module.rules, {"row": 15, "col": 0}) with input.regal.file.lines as lines
	ast.ref_to_string(r3.head.ref) == "rule3"
}

# regal ignore:rule-length
test_find_locals_at_location if {
	policy := `package p

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
`
	module := regal.parse_module("p.rego", policy)
	lines := split(policy, "\n")

	r1 := location.find_locals(module.rules, {"row": 6, "col": 1}) with input as module
		with input.regal.file.lines as lines
	r1 == set()

	r2 := location.find_locals(module.rules, {"row": 6, "col": 10}) with input as module
		with input.regal.file.lines as lines
	r2 == {"x"}

	r3 := location.find_locals(module.rules, {"row": 10, "col": 1}) with input as module
		with input.regal.file.lines as lines
	r3 == {"a", "b"}

	r4 := location.find_locals(module.rules, {"row": 10, "col": 6}) with input as module
		with input.regal.file.lines as lines
	r4 == {"a", "b", "c"}

	r5 := location.find_locals(module.rules, {"row": 15, "col": 1}) with input as module
		with input.regal.file.lines as lines
	r5 == {"x", "y"}

	r6 := location.find_locals(module.rules, {"row": 16, "col": 1}) with input as module
		with input.regal.file.lines as lines
	r6 == {"x", "y", "z"}
}
