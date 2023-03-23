package regal.rules.rules

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result
import data.regal.opa.builtins

builtin_names := {builtin | some builtin, _  in builtins}

# METADATA
# title: rule-shadows-builtin
# description: Rule name shadows built-in
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/rule-shadows-builtin
# custom:
#   category: rules
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	rule.head.name in builtin_names

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}
