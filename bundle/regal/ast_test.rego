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

		[l, m, n] := [1, 2, 3]

		[o, [p, _]] := [1, [2, 1]]

		some _, [q, r] in [["foo", "bar"], [1, 2]]

		{"x": s} := {"x": 1}

		some [t] in [[1]]
	}
	`

	module := regal.parse_module("p.rego", policy)

	vars := ast.find_vars(module.rules)

	names := [name |
		some var in vars
		var.type == "var"
		name := var.value
	]

	names == ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t"]
}

# https://github.com/StyraInc/regal/issues/168
test_function_decls_multiple_same_name if {
	policy := `package p

	import future.keywords.if

	f(x) := x if true
	f(y) := y if false
	`
	module := regal.parse_module("p.rego", policy)
	custom := ast.function_decls(module.rules)

	# we only need to assert there wasn't a conflict in the above
	# call, not what value was returned
	is_object(custom)
}
