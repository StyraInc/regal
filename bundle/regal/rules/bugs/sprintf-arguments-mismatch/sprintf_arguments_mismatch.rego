# METADATA
# description: Mismatch in `sprintf` arguments count
package regal.rules.bugs["sprintf-arguments-mismatch"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

# METADATA
# description: Missing capability for built-in `sprintf`
# custom:
#   severity: none
notices contains result.notice(rego.metadata.chain()) if not "sprintf" in object.keys(config.capabilities.builtins)

# METADATA
# description: |
#   Count the number of distinct arguments ("verbs", denoted by %) in the string argument,
#   compare it to the number of items in the array (if known), and flag when the numbers
#   don't match
report contains violation if {
	some rule_index, fn
	ast.function_calls[rule_index][fn].name == "sprintf"

	# this could come either from a term directly (the common case):
	#     sprintf("%d", [1])
	# or a variable (quite uncommon):
	#     sprintf(format, [1])
	# in the latter case, we try to "resolve" the `format` value by checking if
	# it was assigned in the scope. we do however only do this one level up, and
	# this rule can definitely miss more advanced things, like re-assignemt from
	# another variable. tbh, that's a waste of time. what we should make sure is
	# to not report anything erroneously.
	format_term := _first_arg_value(rule_index, fn.args[0])

	fn.args[1].type == "array" # can only check static arrays, not vars

	values_in_arr := count(fn.args[1].value)
	str_no_escape := replace(format_term.value, "%%", "") # don't include '%%' as it's used to "escape" %
	values_in_str := strings.count(str_no_escape, "%") - _repeated_explicit_argument_indexes(str_no_escape)

	values_in_str != values_in_arr

	violation := result.fail(rego.metadata.chain(), result.ranged_location_between(fn.args[0], regal.last(fn.args)))
}

# see: https://pkg.go.dev/fmt#hdr-Explicit_argument_indexes
# each distinct explicit argument index should only contribute one value to the
# values array. this calculates the number to subtract from the total expected
# number of values based on the number of eai's occurring more than once
_repeated_explicit_argument_indexes(str) := sum([n |
	some eai in _unique_explicit_arguments(str)
	n := strings.count(str, eai) - 1
])

_unique_explicit_arguments(str) := {eai | some eai in regex.find_n(`%\[\d\]`, str, -1)}

_first_arg_value(_, term) := term if term.type == "string"

_first_arg_value(rule_index, term) := found if {
	term.type == "var"

	trow := util.to_location_object(term.location).row

	found := [rhs |
		some expr in ast.exprs[to_number(rule_index)]

		util.to_location_object(expr.location).row < trow

		[lhs, rhs] := ast.assignment_terms(expr)
		lhs.type == "var"
		lhs.value == term.value
		rhs.type == "string"
	][0]
}
