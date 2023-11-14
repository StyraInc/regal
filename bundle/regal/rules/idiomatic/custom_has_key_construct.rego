# METADATA
# description: Custom function may be replaced by `in` and `object.keys`
package regal.rules.idiomatic["custom-has-key-construct"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.capabilities
import data.regal.result

# METADATA
# description: Missing capability for built-in function `object.keys`
# custom:
#   severity: warning
notices contains result.notice(rego.metadata.chain()) if not capabilities.has_object_keys

report contains violation if {
	some rule in ast.functions

	arg_names := ast.function_arg_names(rule)

	count(rule.body) == 1

	terms := rule.body[0].terms

	terms[0].value[0].type == "var"
	terms[0].value[0].value == "eq"

	[_, ref] := normalize_eq_terms(terms)

	ref.value[0].type == "var"
	ref.value[0].value in arg_names
	ref.value[1].type == "var"
	ref.value[1].value in arg_names

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

# METADATA
# description: Normalize var to always always be on the left hand side
normalize_eq_terms(terms) := [terms[1], terms[2]] if {
	terms[1].type == "var"
	startswith(terms[1].value, "$")
	terms[2].type == "ref"
}

normalize_eq_terms(terms) := [terms[2], terms[1]] if {
	terms[1].type == "ref"
	terms[2].type == "var"
	startswith(terms[2].value, "$")
}
