package regal.ast_test

import rego.v1

import data.regal.ast

test_rule_head_locations if {
	policy := `package policy

import rego.v1

default allow := false

allow if true

reasons contains "foo"
reasons contains "bar"

default my_func(_) := false
my_func(1) := true

ref_rule[foo] := true if {
	some foo in [1,2,3]
}
`

	result := ast.rule_head_locations with input as regal.parse_module("p.rego", policy)

	result == {
		"data.policy.allow": {{"col": 9, "row": 5}, {"col": 1, "row": 7}},
		"data.policy.reasons": {{"col": 1, "row": 9}, {"col": 1, "row": 10}},
		"data.policy.my_func": {{"col": 9, "row": 12}, {"col": 1, "row": 13}},
		"data.policy.ref_rule": {{"col": 1, "row": 15}},
	}
}
