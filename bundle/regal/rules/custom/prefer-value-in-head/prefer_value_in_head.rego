# METADATA
# description: Prefer value in rule head
package regal.rules.custom["prefer-value-in-head"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	cfg := config.for_rule("custom", "prefer-value-in-head")

	some rule in input.rules

	var := _var_in_head(rule)
	last := regal.last(rule.body)

	last.terms[0].value[0].type == "var"
	last.terms[0].value[0].value in {"eq", "assign"}
	last.terms[1].type == "var"
	last.terms[1].value == var

	not _scalar_fail(cfg, last.terms[2], ast.scalar_types)

	violation := result.fail(rego.metadata.chain(), result.location(last))
}

_var_in_head(rule) := rule.head.value.value if rule.head.value.type == "var"

_var_in_head(rule) := rule.head.key.value if {
	not rule.head.value
	rule.head.key.type == "var"
}

_scalar_fail(cfg, term, scalar_types) if {
	cfg["only-scalars"] == true
	not term.type in scalar_types
}
