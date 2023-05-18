package regal.rules.rules

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.opa.builtins
import data.regal.result

builtin_names := {builtin | some builtin, _ in builtins}

# METADATA
# title: rule-shadows-builtin
# description: Rule name shadows built-in
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/rule-shadows-builtin
# custom:
#   category: rules
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	rule.head.name in builtin_names

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}

# METADATA
# title: avoid-get-and-list-prefix
# description: Avoid get_ and list_ prefix for rules and functions
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/avoid-get-and-list-prefix
# custom:
#   category: rules
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	strings.any_prefix_match(rule.head.name, {"get_", "list_"})

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}
