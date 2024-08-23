# METADATA
# description: Prefer default assignment over fallback else
package regal.rules.style["default-over-else"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result

cfg := config.for_rule("style", "default-over-else")

report contains violation if {
	some rule in considered_rules

	# walking is expensive but necessary here, since there could be
	# any number of `else` clauses nested below. no need to traverse
	# the rule if there isn't a single `else` present though!
	walk(rule["else"], [_, value])

	else_head := value.head

	not value.body

	ast.is_constant(else_head.value)

	violation := result.fail(rego.metadata.chain(), result.location(else_head))
}

considered_rules := input.rules if cfg["prefer-default-functions"] == true

considered_rules := ast.rules if not cfg["prefer-default-functions"]
