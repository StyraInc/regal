# METADATA
# description: Prefer `some .. in` for iteration
package regal.rules.style["prefer-some-in-iteration"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

cfg := config.for_rule("style", "prefer-some-in-iteration")

report contains violation if {
	some i, rule in input.rules

	not possible_top_level_iteration(rule)

	walk(rule, [path, value])

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
	not invalid_some_context(input.rules[i], path)

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
possible_top_level_iteration(rule) if {
	not rule.body
	rule.head.value.type == "ref"
}

# don't recommend `some .. in` if iteration occurs inside of arrays, objects, or sets
invalid_some_context(rule, path) if {
	some p in util.all_paths(path)

	node := object.get(rule, p, [])

	impossible_some(node)
}

# don't recommend `some .. in` if iteration occurs inside of a
# function call args list, like `startswith(input.foo[_], "foo")`
# this should honestly be a rule of its own, I think, but it's
# not _directly_ replaceable by `some .. in`, so we'll leave it
# be here
invalid_some_context(rule, path) if {
	some p in util.all_paths(path)

	node := object.get(rule, p, [])

	node.terms[0].type == "ref"
	node.terms[0].value[0].type == "var"
	node.terms[0].value[0].value in ast.all_function_names # regal ignore:external-reference
	not node.terms[0].value[0].value in ast.operators # regal ignore:external-reference
}

# if previous node is of type call, also don't recommend `some .. in`
invalid_some_context(rule, path) if object.get(rule, array.slice(path, 0, count(path) - 2), {}).type == "call"

impossible_some(node) if node.type in {"array", "object", "set"}

impossible_some(node) if node.key

# technically this is not an _impossible_ some, as we could replace e.g. `"x" == input[_]`
# with `some "x" in input`, but that'd be an `unnecessary-some` violation as `"x" in input`
# would be the correct way to express that
impossible_some(node) if {
	node.terms[0].value[0].type == "var"
	node.terms[0].value[0].value in {"eq", "equal"}

	some term in node.terms
	term.type in ast.scalar_types # regal ignore:external-reference
}
