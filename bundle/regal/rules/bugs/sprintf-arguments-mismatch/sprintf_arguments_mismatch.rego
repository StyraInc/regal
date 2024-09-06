# METADATA
# description: Mismatch in `sprintf` arguments count
package regal.rules.bugs["sprintf-arguments-mismatch"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result

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
	some fn
	ast.function_calls[_][fn].name == "sprintf"

	fn.args[1].type == "array" # can only check static arrays, not vars

	values_in_arr := count(fn.args[1].value)
	str_no_escape := replace(fn.args[0].value, "%%", "") # don't include '%%' as it's used to "escape" %
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
