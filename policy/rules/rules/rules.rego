package regal.rules.rules

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal
import data.builtins

builtin_names := {builtin | some builtin, _  in builtins}

# METADATA
# title: STY-RULES-001
# description: Rule name shadows built-in
# related_resources:
# - https://docs.styra.com/regal/rules/sty-rules-001
violation contains msg if {
	some rule in input.rules
	rule.head.name in builtin_names

	msg := regal.fail(rego.metadata.rule(), {})
}
