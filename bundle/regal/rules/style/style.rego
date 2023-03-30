package regal.rules.style

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result
import data.regal.util

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	not util.is_snake_case(rule.head.name)

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	rule["default"] == true
	not util.is_snake_case(rule.head.name)

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	some expr in rule.body
	some symbol in expr.terms.symbols

	# allow { some camelCase }
	is_string(symbol.value)
	not util.is_snake_case(symbol.value)

	violation := result.fail(rego.metadata.rule(), result.location(symbol))
}

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	some expr in rule.body

	# allow { camelCase := "wrong" }
	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "assign"
	expr.terms[1].type == "var"

	not util.is_snake_case(expr.terms[1].value)

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms[1]))
}

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	some expr in rule.body
	some symbol in expr.terms.symbols

	# allow { some camelCase in input }
	symbol.type == "call"
	symbol.value[1].type == "var"

	not util.is_snake_case(symbol.value[1].value)

	violation := result.fail(rego.metadata.rule(), result.location(symbol.value[1]))
}

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	some expr in rule.body
	some symbol in expr.terms.symbols

	# allow { some x, camelCase in input }
	symbol.type == "call"
	symbol.value[2].type == "var"

	not util.is_snake_case(symbol.value[2].value)

	violation := result.fail(rego.metadata.rule(), result.location(symbol.value[2]))
}

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	some expr in rule.body

	# allow { every camelCaseKey, value in input {...}}
	expr.terms.domain
	expr.terms.key.type == "var"

	not util.is_snake_case(expr.terms.key.value)

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms.key))
}

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	some expr in rule.body

	# allow { every x, camelCase in input {...}}
	expr.terms.domain
	expr.terms.value.type == "var"

	not util.is_snake_case(expr.terms.value.value)

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms.value))
}

# TODO: scan doesn't currently go into the body of
# `every` expressions
