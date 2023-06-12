package regal.rules.style

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# METADATA
# title: use-in-operator
# description: Use in to check for membership
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/use-in-operator
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some expr in eq_exprs

	expr.terms[1].type in {"array", "boolean", "object", "null", "number", "set", "string", "var"}
	expr.terms[2].type == "ref"

	last := regal.last(expr.terms[2].value)

	last.type == "var"
	startswith(last.value, "$")

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms[2].value[0]))
}

# METADATA
# title: use-in-operator
# description: Use in to check for membership
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/use-in-operator
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some expr in eq_exprs

	expr.terms[1].type == "ref"
	expr.terms[2].type in {"array", "boolean", "object", "null", "number", "set", "string", "var"}

	last := regal.last(expr.terms[1].value)

	last.type == "var"
	startswith(last.value, "$")

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms[1].value[0]))
}

eq_exprs contains expr if {
	config.for_rule({"category": "style", "title": "use-in-operator"}).level != "ignore"

	some rule in input.rules
	some expr in rule.body

	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "equal"
}
