package regal.rules.bugs

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# METADATA
# title: top-level-iteration
# description: Iteration in top-level assignment
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/top-level-iteration
# custom:
#   category: bugs
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules

	rule.head.value.type == "ref"

	last := regal.last(rule.head.value.value)
	last.type == "var"

	illegal_value_ref(last.value)

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}

_rule_names := {name | name := input.rules[_].head.name}

# regal ignore:external-reference
illegal_value_ref(value) if not value in _rule_names
