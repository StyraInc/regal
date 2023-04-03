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

		[l, m, n] := jwt.decode("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.Et9HFtf9R3GEMA0IICOfFMVXY7kkTX1wr4qCyhIf58U")
	}
	`

	module := regal.parse_module("p.rego", policy)

	vars := ast.find_vars(module.rules)

	names := [name |
		some var in vars
		var.type == "var"
		name := var.value
	]

	print(names)

	names == ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n"]
}
