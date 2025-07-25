# METADATA
# description: Assignment can be deferred
package regal.rules.performance["defer-assignment"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some i, rule in input.rules
	some j, expr in rule.body

	[var, rhs] := ast.assignment_terms(expr.terms)

	not _ref_with_vars(rhs)

	# for now, only simple var assignment counts.. later we can
	# consider checking the contents of arrays here
	var.type == "var"

	next := rule.body[j + 1]

	not ast.is_assignment(next.terms[0])
	not ast.var_in_head(rule.head, var.value)
	not _var_value_used_in_expression(var.value, next)
	not _iteration_expression(next.terms)
	not _print_call(next.terms)

	violation := result.fail(rego.metadata.chain(), result.location(expr))
}

_ref_with_vars(node) if {
	node.type == "ref"

	some i, term in node.value

	i > 0
	term.type == "var"
}

_var_value_used_in_expression(value, expr) if {
	not expr.terms.symbols

	is_array(expr.terms)

	some term in expr.terms

	walk(term, [_, node])

	node.type == "var"
	node.value == value
}

_var_value_used_in_expression(value, expr) if {
	some w in expr.with

	walk(w, [_, node])

	node.type == "var"
	node.value == value
}

_var_value_used_in_expression(value, expr) if {
	# `not x`
	is_object(expr.terms)

	expr.terms.type == "var"
	expr.terms.value == value
} else if {
	# `not x.y`
	is_object(expr.terms)

	some term in expr.terms.value

	walk(term, [_, node])

	node.type == "var"
	node.value == value
}

# while not technically checking of use here:
# the next expression having symbols indicate iteration, and
# we don't want to defer assignment into a loop
_iteration_expression(terms) if terms.symbols

# likewise with every
_iteration_expression(terms) if terms.domain

# and walk
_iteration_expression(terms) if {
	terms[0].value[0].type == "var"
	terms[0].value[0].value == "walk"
}

# regal ignore:narrow-argument
_print_call(terms) if {
	terms[0].value[0].type == "var"
	terms[0].value[0].value == "print"
}
