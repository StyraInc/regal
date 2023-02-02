package regal.rules.style

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal

# METADATA
# title: STY-STYLE-001
# description: Prefer snake_case for names
# related_resources:
# - https://docs.styra.com/regal/rules/sty-style-001
violation contains msg if {
	some rule in input.rules
	not regal.is_snake_case(rule.head.name)

	msg := regal.fail(rego.metadata.rule(), {})
}

# METADATA
# title: STY-STYLE-001
# description: Prefer snake_case for names
# related_resources:
# - https://docs.styra.com/regal/rules/sty-style-001
violation contains msg if {
	some rule in input.rules
	some expr in rule.body
	some symbol in expr.terms.symbols

	# allow { some camelCase }
	is_string(symbol.value)
	not regal.is_snake_case(symbol.value)

	msg := regal.fail(rego.metadata.rule(), {})
}

# METADATA
# title: STY-STYLE-001
# description: Prefer snake_case for names
# related_resources:
# - https://docs.styra.com/regal/rules/sty-style-001
violation contains msg if {
	some rule in input.rules
	some expr in rule.body

	# allow { camelCase := "wrong" }
	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "assign"
	expr.terms[1].type == "var"

	not regal.is_snake_case(expr.terms[1].value)

	msg := regal.fail(rego.metadata.rule(), {})
}

# METADATA
# title: STY-STYLE-001
# description: Prefer snake_case for names
# related_resources:
# - https://docs.styra.com/regal/rules/sty-style-001
violation contains msg if {
	some rule in input.rules
	some expr in rule.body
	some symbol in expr.terms.symbols

	# allow { some camelCase in input }
	symbol.type == "call"
	symbol.value[1].type == "var"

	not regal.is_snake_case(symbol.value[1].value)

	msg := regal.fail(rego.metadata.rule(), {})
}

# METADATA
# title: STY-STYLE-001
# description: Prefer snake_case for names
# related_resources:
# - https://docs.styra.com/regal/rules/sty-style-001
violation contains msg if {
	some rule in input.rules
	some expr in rule.body
	some symbol in expr.terms.symbols

	# allow { some x, camelCase in input }
	symbol.type == "call"
	symbol.value[2].type == "var"

	not regal.is_snake_case(symbol.value[2].value)

	msg := regal.fail(rego.metadata.rule(), {})
}

# METADATA
# title: STY-STYLE-001
# description: Prefer snake_case for names
# related_resources:
# - https://docs.styra.com/regal/rules/sty-style-001
violation contains msg if {
	some rule in input.rules
	some expr in rule.body

	# allow { every camelCaseKey, value in input {...}}
	expr.terms.domain
	expr.terms.key.type == "var"

	not regal.is_snake_case(expr.terms.key.value)

	msg := regal.fail(rego.metadata.rule(), {})
}

# METADATA
# title: STY-STYLE-001
# description: Prefer snake_case for names
# related_resources:
# - https://docs.styra.com/regal/rules/sty-style-001
violation contains msg if {
	some rule in input.rules
	some expr in rule.body

	# allow { every x, camelCase in input {...}}
	expr.terms.domain
	expr.terms.value.type == "var"

	not regal.is_snake_case(expr.terms.value.value)

	msg := regal.fail(rego.metadata.rule(), {})
}

# TODO: scan doesn't currently go into the body of
# `every` expressions
