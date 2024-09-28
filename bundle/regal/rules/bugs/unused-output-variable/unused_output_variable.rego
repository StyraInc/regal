# METADATA
# description: Unused output variable
package regal.rules.bugs["unused-output-variable"]

import rego.v1

import data.regal.ast
import data.regal.result

# METADATA
# description: |
#   The `report` set contains unused output vars from `some` declarations. Using a
#   stricter ruleset than OPA, Regal considers an output var "unused" if it's used
#   only once in a ref, as that usage may just as well be replaced by a wildcard.
#   ```
#   some y
#   x := data.foo.bar[y]
#   # y not used later
#   ```
#   Would better be written as:
#   ```
#   some x in data.foo.bar
#   ```
report contains violation if {
	some rule_index
	var_refs := _ref_vars[rule_index][_]

	count(var_refs) == 1

	some var in var_refs

	not _var_in_head(input.rules[to_number(rule_index)].head, var.value)
	not _var_in_call(ast.function_calls, rule_index, var.value)
	not _ref_base_vars[rule_index][var.value]

	# this is by far the most expensive condition to check, so only do
	# so when all other conditions apply
	ast.is_output_var(input.rules[to_number(rule_index)], var)

	violation := result.fail(rego.metadata.chain(), result.location(var))
}

_ref_vars[rule_index][var.value] contains var if {
	some rule_index
	var := ast.found.vars[rule_index].ref[_]

	not startswith(var.value, "$")
}

# "a" in "a[foo]", and not foo
_ref_base_vars[rule_index][term.value] contains term if {
	some rule_index
	term := ast.found.refs[rule_index][_].value[0]

	term.type == "var"
	not startswith(term.value, "$")
}

_var_in_head(head, name) if head.value.value == name

_var_in_head(head, name) if {
	some var in ast.find_term_vars(head.value.value)

	var.value == name
}

_var_in_head(head, name) if head.key.value == name

_var_in_head(head, name) if {
	some var in ast.find_term_vars(head.key.value)

	var.value == name
}

_var_in_head(head, name) if {
	some i, var in head.ref
	i > 0
	var.value == name
}

_var_in_call(calls, rule_index, name) if _var_in_arg(calls[rule_index][_].args[_], name)

_var_in_arg(arg, name) if {
	arg.type == "var"
	arg.value == name
}

_var_in_arg(arg, name) if {
	arg.type in {"array", "object", "set"}

	some var in ast.find_term_vars(arg)

	var.value == name
}
