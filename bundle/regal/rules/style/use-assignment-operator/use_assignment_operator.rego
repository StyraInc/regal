# METADATA
# description: Prefer := over = for assignment
package regal.rules.style["use-assignment-operator"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	# foo = "bar"
	# default foo = "bar
	# foo(bar) = "baz"
	# default foo(_) = "bar
	some rule in input.rules
	not rule.head.assign
	not rule.head.key
	not ast.implicit_boolean_assignment(rule)
	not ast.is_chained_rule_body(rule, input.regal.file.lines)

	loc := result.location(rule)

	violation := result.fail(rego.metadata.chain(), object.union(loc, {"location": {"col": _eq_col(loc)}}))
}

report contains violation if {
	# foo["bar"] = "baz"
	some rule in ast.rules
	rule.head.key
	rule.head.value
	not rule.head.assign

	loc := result.location(result.location(rule.head.ref[0]))

	violation := result.fail(rego.metadata.chain(), object.union(loc, {"location": {"col": _eq_col(loc)}}))
}

report contains violation if {
	some rule in input.rules

	# walking is expensive but necessary here, since there could be
	# any number of `else` clauses nested below. no need to traverse
	# the rule if there isn't a single `else` present though!

	# NOTE: the same logic is used in default-over-else
	# we should consider having a helper function to return
	# all else clauses, for a given rule, as potentially that
	# would be cached on the second invocation of the function
	walk(rule["else"], [_, value])

	loc := result.location(value.head)

	# extract the text from location to see if '=' is used for
	# assignment
	regex.match(`else\s*=`, loc.location.text)

	violation := result.fail(rego.metadata.chain(), object.union(loc, {"location": {"col": _eq_col(loc)}}))
}

_eq_col(loc) := max([0, indexof(loc.location.text, "=")]) + 1
