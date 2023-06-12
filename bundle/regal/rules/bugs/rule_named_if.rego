package regal.rules.bugs

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# METADATA
# title: rule-named-if
# description: Rule named "if"
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/rule-named-if
# custom:
#   category: bugs
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	rule.head.name == "if"

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}
