package regal.rules.bugs

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.opa
import data.regal.result

# METADATA
# title: unused-return-value
# description: Non-boolean return value unused
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/unused-return-value
# custom:
#   category: bugs
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	some expr in rule.body

	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"

	ref_name := expr.terms[0].value[0].value
	ref_name in ast.builtin_names

	opa.builtins[ref_name].result.type != "boolean"

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms[0]))
}
