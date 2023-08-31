# METADATA
# description: Prefer `some .. in` for iteration
package regal.rules.style["prefer-some-in-iteration"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

cfg := config.for_rule("style", "prefer-some-in-iteration")

report contains violation if {
	some rule in input.rules

	node := filter_top_level_ref(rule)

	# regal ignore:function-arg-return,unused-return-value
	walk(node, [_, value])

	value.type == "ref"

	vars_in_ref := ast.find_ref_vars(value)

	count(vars_in_ref) > 0

	num_output_vars := count([var |
		some var in vars_in_ref

		# we don't need the location of each var, but using the first
		# ref will do, and will hopefully help with caching the result
		ast.is_output_var(rule, var, value.location)
	])

	num_output_vars != 0
	num_output_vars < cfg["ignore-nesting-level"]

	violation := result.fail(rego.metadata.chain(), result.location(value))
}

# don't walk top level iteration refs:
# https://docs.styra.com/regal/rules/bugs/top-level-iteration
filter_top_level_ref(rule) := rule.body if {
	rule.head.value.type == "ref"
} else := rule
