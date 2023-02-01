package regal.rules.assignment

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal

# METADATA
# title: STY-ASSIGN-001
# description: Prefer := over = for assignment
# related_resources:
# - https://docs.styra.com/regal/rules/sty-assign-001
violation contains msg if {
	some rule in input.rules
	rule["default"] == true
	not rule.head.assign

	msg := regal.fail(rego.metadata.rule(), {})
}

# METADATA
# title: STY-ASSIGN-001
# description: Prefer := over = for assignment
# related_resources:
# - https://docs.styra.com/regal/rules/sty-assign-001
violation contains msg if {
	some rule in input.rules
	rule.head.key
	rule.head.value
	not rule.head.assign

	msg := regal.fail(rego.metadata.rule(), {})
}
