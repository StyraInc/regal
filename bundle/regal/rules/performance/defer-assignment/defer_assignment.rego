# METADATA
# description: Assignment can be deferred
package regal.rules.performance["defer-assignment"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some i, rule in input.rules
	some j, expr in rule.body

	[var, rhs] := ast.assignment_terms(expr)

	not _ref_with_vars(rhs)

	# for now, only simple var assignment counts.. later we can
	# consider checking the contents of arrays here
	var.type == "var"

	next := rule.body[j + 1]

	not ast.is_assignment(next)
	not ast.var_in_head(rule.head, var.value)
	not _var_used_in_expression(var, next)
	not _iteration_expression(next)
	not _print_call(next)

	violation := result.fail(rego.metadata.chain(), result.location(expr))
}

_ref_with_vars(node) if {
	node.type == "ref"

	some i, term in node.value

	i > 0
	term.type == "var"
}

_var_used_in_expression(var, expr) if {
	not expr.terms.symbols

	is_array(expr.terms)

	some term in expr.terms

	walk(term, [_, value])

	value.type == "var"
	value.value == var.value
}

_var_used_in_expression(var, expr) if {
	some w in expr["with"]

	walk(w, [_, value])

	value.type == "var"
	value.value == var.value
}

_var_used_in_expression(var, expr) if {
	# `not x`
	is_object(expr.terms)

	expr.terms.type == "var"
	expr.terms.value == var.value
} else if {
	# `not x.y`
	is_object(expr.terms)

	some term in expr.terms.value

	walk(term, [_, value])

	value.type == "var"
	value.value == var.value
}

# while not technically checking of use here:
# the next expression having symbols indicate iteration, and
# we don't want to defer assignment into a loop
_iteration_expression(expr) if expr.terms.symbols

# likewise with every
_iteration_expression(expr) if expr.terms.domain

# and walk
_iteration_expression(expr) if {
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "walk"
}

_print_call(expr) if {
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "print"
}
