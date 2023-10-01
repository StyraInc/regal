# METADATA
# description: Prefer default assignment over fallback else
package regal.rules.style["default-over-else"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

cfg := config.for_rule("style", "default-over-else")

report contains violation if {
	some rule in considered_rules

	# walking is expensive but necessary here, since there could be
	# any number of `else` clauses nested below. no need to traverse
	# the rule if there isn't a single `else` present though!
	rule["else"]

	walk(rule, [_, value])

	# quoting is needed as `else` is a keyword
	else_body := value["else"].body
	else_head := value["else"].head

	# we don't know for sure, but if all that's in the body is an empty
	# `true`, it's likely an implicit body (i.e. one not printed)
	count(else_body) == 1
	else_body[0].terms.type == "boolean"
	else_body[0].terms.value == true

	ast.is_constant(else_head.value)

	violation := result.fail(rego.metadata.chain(), result.location(else_head))
}

considered_rules := input.rules if cfg["prefer-default-functions"] == true

considered_rules := [rule | some rule in input.rules; not rule.head.args] if not cfg["prefer-default-functions"]
