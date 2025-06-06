# METADATA
# description: Prefer value in rule head
package regal.rules.custom["prefer-value-in-head"]

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	some rule in input.rules

	var := _var_in_head(rule.head)
	terms := regal.last(rule.body).terms

	terms[0].value[0].type == "var"
	terms[0].value[0].value in {"eq", "assign"}
	terms[1].type == "var"
	terms[1].value == var

	not _scalar_fail(terms[2].type, ast.scalar_types)
	not _excepted_var_name(var)

	violation := result.fail(rego.metadata.chain(), result.location(terms[2]))
}

_var_in_head(head) := head.value.value if head.value.type == "var"

_var_in_head(head) := head.key.value if {
	not head.value
	head.key.type == "var"
}

_scalar_fail(term_type, scalar_types) if {
	config.rules.custom["prefer-value-in-head"]["only-scalars"] == true
	not term_type in scalar_types
}

_excepted_var_name(name) if name in config.rules.custom["prefer-value-in-head"]["except-var-names"]
