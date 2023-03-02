package regal.rules.rules

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal
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
	regal.rule_config(rego.metadata.rule()).enabled == true

	some rule in input.rules
	rule.head.name in builtin_names

	violation := regal.fail(rego.metadata.rule(), {})
}
