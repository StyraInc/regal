# METADATA
# description: Prefer `some .. in` for iteration
package regal.rules.style["prefer-some-in-iteration"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result

cfg := config.for_rule("style", "prefer-some-in-iteration")

report contains violation if {
	some rule in input.rules

	node := filter_top_level_ref(rule)

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

	not except_sub_attribute(value.value)

	violation := result.fail(rego.metadata.chain(), result.location(value))
}

except_sub_attribute(ref) if {
	cfg["ignore-if-sub-attribute"] == true
	has_sub_attribute(ref)
}

has_sub_attribute(ref) if {
	last_var_pos := regal.last([i |
		some i, part in ref
		part.type == "var"
	])
	last_var_pos < count(ref) - 1
}

# don't walk top level iteration refs:
# https://docs.styra.com/regal/rules/bugs/top-level-iteration
filter_top_level_ref(rule) := rule.body if {
	rule.head.value.type == "ref"
} else := rule
