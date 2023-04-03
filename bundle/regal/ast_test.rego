package regal.ast_test

import future.keywords.if
import future.keywords.in

import data.regal.ast

test_find_vars if {
	policy := `
	package p

	import future.keywords

	global := "foo"

	allow if {
		a := global
		b := [c | c := input[x]] # can't capture x

		every d in input {
			d == "foo"
		}

		every e, f in input.bar {
			e == f
		}

		some g, h
		input.bar[g][h]
		some i in input
		some j, k in input
	}
	`

	module := regal.parse_module("p.rego", policy)

    vars := ast.find_vars(module.rules)

	names := [name |
		some var in vars
		var.type == "var"
		name := var.value
	]

	names == ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
}
