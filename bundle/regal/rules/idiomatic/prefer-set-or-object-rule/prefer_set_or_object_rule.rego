# METADATA
# description: Prefer set or object rule over comprehension
package regal.rules.idiomatic["prefer-set-or-object-rule"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in ast.rules

	rule.head.value.type in {"setcomprehension", "objectcomprehension"}
	not rule.body

	# Ignore simple conversions from array to set
	not _is_array_conversion(rule.head.value)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

# {s | some s in arr}
_is_array_conversion(value) if {
	value.type == "setcomprehension"
	value.value.term.type == "var"

	var := value.value.term.value
	body := value.value.body

	count(body) == 1

	symbols := body[0].terms.symbols

	count(symbols) == 1

	symbols[0].type == "call"
	symbols[0].value[0].type == "ref"
	symbols[0].value[0].value[0].type == "var"
	symbols[0].value[0].value[0].value == "internal"
	symbols[0].value[0].value[1].value == "member_2"
	symbols[0].value[1].type == "var"
	symbols[0].value[1].value == var
}

# {s | s := arr[_]}
# or
# {s | s := arr[_].foo}
# or
# {s | s := arr[_].foo[_]}
_is_array_conversion(value) if {
	value.type == "setcomprehension"
	value.value.term.type == "var"

	var := value.value.term.value
	body := value.value.body

	count(body) == 1

	[lhs, rhs] := ast.assignment_terms(body[0])

	# Assignment to comprehension variable
	lhs.type == "var"
	lhs.value == var

	# On the right hand side a ref with at least one wildcard
	rhs.type == "ref"
	rhs.value[0].type == "var"

	some ref_val in rhs.value

	ref_val.type == "var"
	startswith(ref_val.value, "$")
}
