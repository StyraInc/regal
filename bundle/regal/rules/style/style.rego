package regal.rules.style

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
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

	some var in ast.find_vars(input.rules)
	not util.is_snake_case(var.value)

	violation := result.fail(rego.metadata.rule(), result.location(var))
}

# METADATA
# title: use-in-operator
# description: Use in to check for membership
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/use-in-operator
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some expr in eq_exprs

	expr.terms[1].type in {"array", "boolean", "object", "null", "number", "set", "string", "var"}
	expr.terms[2].type == "ref"

	last := expr.terms[2].value[count(expr.terms[2].value) - 1]

	last.type == "var"
	startswith(last.value, "$")

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms[2].value[0]))
}

# METADATA
# title: use-in-operator
# description: Use in to check for membership
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/use-in-operator
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some expr in eq_exprs

	expr.terms[1].type == "ref"
	expr.terms[2].type in {"array", "boolean", "object", "null", "number", "set", "string", "var"}

	last := expr.terms[1].value[count(expr.terms[1].value) - 1]

	last.type == "var"
	startswith(last.value, "$")

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms[1].value[0]))
}

eq_exprs contains expr if {
	config.for_rule({"category": "style", "title": "use-in-operator"}).enabled == true

	some rule in input.rules
	some expr in rule.body

	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "equal"
}

# METADATA
# title: line-length
# description: Line too long
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/line-length
# custom:
#   category: style
report contains violation if {
	cfg := config.for_rule(rego.metadata.rule())

	cfg.enabled == true

	some i, line in input.regal.file.lines

	line_length := count(line)
	line_length > cfg["max-line-length"]

	violation := result.fail(
		rego.metadata.rule(),
		{"location": {
			"file": input.regal.file.name,
			"row": i + 1,
			"col": line_length,
			"text": input.regal.file.lines[i],
		}},
	)
}
