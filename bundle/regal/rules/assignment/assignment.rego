package regal.rules.assignment

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal

# METADATA
# title: use-assignment-operator
# description: Prefer := over = for assignment
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/use-assignment-operator
# custom:
#   category: assignment
report contains violation if {
    regal.rule_config(rego.metadata.rule()).enabled == true

	some rule in input.rules
	rule["default"] == true
	not rule.head.assign

	violation := regal.fail(rego.metadata.rule(), {})
}

# METADATA
# title: assignment-operator
# description: Prefer := over = for assignment
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/use-assignment-operator
# custom:
#   category: assignment
report contains violation if {
	regal.rule_config(rego.metadata.rule()).enabled == true

	some rule in input.rules
	rule.head.key
	rule.head.value
	not rule.head.assign

	violation := regal.fail(rego.metadata.rule(), {})
}
