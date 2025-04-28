# METADATA
# description: Impossible `not` condition
package regal.rules.bugs["impossible-not"]

import data.regal.ast
import data.regal.result
import data.regal.util

# note: not ast.package_path as we want the "data" component here
_package_path := [part.value | some part in input["package"].path]

_multivalue_rules contains path if {
	some rule in ast.rules

	rule.head.key
	not rule.head.value

	# ignore general ref head rules for now
	every path in array.slice(rule.head.ref, 1, count(rule.head.ref)) {
		path.type == "string"
	}

	path := concat(".", array.concat(_package_path, [ref.value | some ref in rule.head.ref]))
}

_negated_refs contains negated_ref if {
	some rule_index, value
	ast.found.expressions[rule_index][value].negated

	# if terms is an array, it's a function call, and most likely not "impossible"
	is_object(value.terms)
	value.terms.type in {"ref", "var"}

	ref := _var_to_ref(value.terms)

	# for now, ignore ref if it has variable components
	every path in util.rest(ref) {
		path.type == "string"
	}

	rule := input.rules[to_number(rule_index)]

	# ignore negated local vars
	not ref[0].value in ast.function_arg_names(rule)
	not ref[0].value in {var.value |
		some var in ast.find_vars_in_local_scope(rule, util.to_location_object(value.location))
	}

	term1 := object.union(ref[0], {"location": util.to_location_object(ref[0].location)})

	negated_ref := {
		"ref": array.concat([term1], util.rest(ref)),
		"resolved_path": _resolve(ref, _package_path, ast.resolved_imports),
	}
}

# METADATA
# description: collects imported symbols, multi-value rules and negated refs
aggregate contains result.aggregate(rego.metadata.chain(), {
	"imported_symbols": ast.resolved_imports,
	"multivalue_rules": _multivalue_rules,
	"negated_refs": _negated_refs,
})

report contains violation if {
	some entry in aggregate
	some negated in entry.aggregate_data.negated_refs

	negated.resolved_path in entry.aggregate_data.multivalue_rules

	loc := object.union(result.location(negated.ref), {"location": {
		"file": entry.aggregate_source.file,
		# note that the "not" isn't present in the AST, so we'll add it manually to the text
		# in the location to try and make it clear where the issue is (as opposed to just
		# printing the ref)
		"text": sprintf("not %s", [_to_string(negated.ref)]),
	}})

	violation := result.fail(rego.metadata.chain(), loc)
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
aggregate_report contains violation if {
	all_multivalue_refs := {{"path": path, "file": entry.aggregate_source.file} |
		some entry in input.aggregate
		some path in entry.aggregate_data.multivalue_rules
	}

	some entry in input.aggregate
	some negated in entry.aggregate_data.negated_refs
	some multivalue_ref in all_multivalue_refs

	# violations in same file handled by non-aggregate "report"
	multivalue_ref.file != entry.aggregate_source.file

	negated.resolved_path == multivalue_ref.path

	loc := object.union(result.location(negated.ref), {"location": {
		"file": entry.aggregate_source.file,
		# note that the "not" isn't present in the AST, so we'll add it manually to the text
		# in the location to try and make it clear where the issue is (as opposed to just
		# printing the ref)
		"text": sprintf("not %s", [_to_string(negated.ref)]),
	}})

	violation := result.fail(rego.metadata.chain(), loc)
}

_var_to_ref(terms) := [terms] if terms.type == "var"
_var_to_ref(terms) := terms.value if terms.type == "ref"

_to_string(ref) := concat(".", [part.value | some part in ref])

_resolve(ref, _, _) := _to_string(ref) if ref[0].value == "data"

# imported symbol
_resolve(ref, _, imported_symbols) := concat(".", resolved) if {
	ref[0].value != "data"

	resolved := array.concat(
		imported_symbols[ref[0].value],
		[part.value | some part in array.slice(ref, 1, count(ref))],
	)
}

# not imported — must be local or package
_resolve(ref, pkg_path, imported_symbols) := concat(".", resolved) if {
	ref[0].value != "data"

	not imported_symbols[ref[0].value]

	resolved := array.concat(
		pkg_path,
		[part.value | some part in ref],
	)
}
