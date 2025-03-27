# METADATA
# description: Prefer default assignment over fallback else
package regal.rules.style["default-over-else"]

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	some rule in _considered_rules

	# walking is expensive but necessary here, since there could be
	# any number of `else` clauses nested below. no need to traverse
	# the rule if there isn't a single `else` present though!
	walk(rule["else"], [_, value])

	not value.body

	else_head := value.head

	ast.is_constant(else_head.value)

	violation := result.fail(rego.metadata.chain(), result.location(else_head))
}

_cfg := config.for_rule("style", "default-over-else")

_considered_rules := input.rules if _cfg["prefer-default-functions"] == true
_considered_rules := ast.rules if not _cfg["prefer-default-functions"]
