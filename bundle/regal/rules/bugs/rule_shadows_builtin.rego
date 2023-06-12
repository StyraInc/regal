package regal.rules.bugs

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

# METADATA
# title: rule-shadows-builtin
# description: Rule name shadows built-in
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/rule-shadows-builtin
# custom:
#   category: bugs
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	rule.head.name in ast.builtin_names

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}
