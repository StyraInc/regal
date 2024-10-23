# METADATA
# description: External reference in function
package regal.rules.style["external-reference"]

import rego.v1

import data.regal.ast
import data.regal.result
import data.regal.util

report contains violation if {
	fn_namespaces := {split(name, ".")[0] | some name in object.keys(ast.all_functions)}

	some fn in ast.functions

	named_args := {arg.value | some arg in fn.head.args; arg.type == "var"}

	head_vars := {v.value | some v in ast.find_vars(fn.head.value)}
	body_vars := {v.value | some v in ast.find_vars(fn.body)}
	else_vars := {v.value | some v in ast.find_vars(fn["else"])}
	own_vars := (body_vars | head_vars) | else_vars

	# note: parens added by opa fmt 🤦
	allowed_refs := (named_args | own_vars) | fn_namespaces

	walk(fn, [path, value])

	value.type == "var"
	not value.value in allowed_refs
	not ast.is_wildcard(value)
	not _function_call_ctx(fn, path)

	violation := result.fail(rego.metadata.chain(), result.location(value))
}

# METADATA
# scope: document
# description: |
#   functions should be able to call other functions and this shouldn't
#   be flagged as there's no way to import other functions via arguments
#   note: this doesn't check for built-in calls or calls to function
#   defined in the same package, as those are already covered by
#   "fn_namespaces" in the report rule
_function_call_ctx(fn, path) if {
	terms_path := array.slice(path, 0, util.last_indexof(path, "terms") + 2)
	next_term_path := array.concat(
		array.slice(terms_path, 0, count(terms_path) - 1), # ["body", 0, "terms", 0] -> ["body", 0, "terms"]
		[regal.last(terms_path) + 1], # 0 -> 1
	)

	# ["body", 0, "terms", 1]

	object.get(fn, next_term_path, null) != null
}

_function_call_ctx(fn, path) if object.get(fn, array.slice(path, 0, count(path) - 4), {}).type == "call"
