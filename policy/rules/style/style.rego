package regal.rules.style

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal

# METADATA
# title: STY-STYLE-001
# description: Prefer snake_case for names
# related_resources:
# - https://docs.styra.com/regal/rules/sty-style-001
violation contains msg if {
    some rule in input.rules
    not regal.is_snake_case(rule.head.name)

	msg := regal.fail(rego.metadata.rule(), {})
}
