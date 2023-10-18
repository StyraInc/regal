# METADATA
# description: Prefer value in rule head
package regal.rules.custom["prefer-value-in-head"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

cfg := config.for_rule("custom", "prefer-value-in-head")

report contains violation if {
	some rule in input.rules

	var := var_in_head(rule)
	last := regal.last(rule.body)

	last.terms[0].value[0].type == "var"
	last.terms[0].value[0].value in {"eq", "assign"}
	last.terms[1].type == "var"
	last.terms[1].value == var

	not configured_exception(cfg, last.terms[2])

	violation := result.fail(rego.metadata.chain(), result.location(last))
}

var_in_head(rule) := rule.head.value.value if rule.head.value.type == "var"

var_in_head(rule) := rule.head.key.value if {
	not rule.head.value
	rule.head.key.type == "var"
}

configured_exception(cfg, term) if {
	cfg["only-scalars"] == true
	term.type in ast.scalar_types
}
