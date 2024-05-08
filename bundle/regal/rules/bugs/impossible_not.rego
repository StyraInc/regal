# METADATA
# description: Impossible `not` condition
package regal.rules.bugs["impossible-not"]

import rego.v1

import data.regal.ast
import data.regal.result

# regal ignore:rule-length
aggregate contains entry if {
	package_path := [part.value | some part in input["package"].path]

	imported_symbols := {symbol: path |
		some _import in input.imports

		_import.path.value[0].value == "data"
		count(_import.path.value) > 1

		symbol := imported_symbol(_import)
		path := [part.value | some part in _import.path.value]
	}

	multivalue_rules := {path |
		some rule in ast.rules

		rule.head.key
		not rule.head.value

		# ignore general ref head rules for now
		every path in array.slice(rule.head.ref, 1, count(rule.head.ref)) {
			path.type == "string"
		}

		path := concat(".", array.concat(package_path, [p |
			some ref in rule.head.ref
			p := ref.value
		]))
	}

	negated_refs := [negated_ref |
		some i, rule in input.rules

		walk(rule, [_, value])

		value.negated

		# if terms is an array, it's a function call, and most likely not "impossible"
		is_object(value.terms)
		value.terms.type in {"ref", "var"}

		ref := var_to_ref(value.terms)

		# for now, ignore ref if it has variable components
		every path in array.slice(ref, 1, count(ref)) {
			path.type == "string"
		}

		# ignore negated local vars
		not ref[0].value in ast.function_arg_names(rule)
		not ref[0].value in {var.value | some var in ast.find_vars_in_local_scope(rule, value.location)}

		negated_ref := {
			"ref": ref,
			"resolved_path": resolve(ref, package_path, imported_symbols),
		}
	]

	entry := result.aggregate(rego.metadata.chain(), {
		"imported_symbols": imported_symbols,
		"multivalue_rules": multivalue_rules,
		"negated_refs": negated_refs,
	})
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
aggregate_report contains violation if {
	all_multivalue_refs := {path |
		some entry in input.aggregate
		some path in entry.aggregate_data.multivalue_rules
	}

	some entry in input.aggregate
	some negated in entry.aggregate_data.negated_refs

	negated.resolved_path in all_multivalue_refs

	loc := object.union(result.location(negated.ref), {"location": {
		"file": entry.aggregate_source.file,
		# note that the "not" isn't present in the AST, so we'll add it manually to the text
		# in the location to try and make it clear where the issue is (as opposed to just
		# printing the ref)
		"text": sprintf("not %s", [to_string(negated.ref)]),
	}})

	violation := result.fail(rego.metadata.chain(), loc)
}

var_to_ref(terms) := [terms] if terms.type == "var"

var_to_ref(terms) := terms.value if terms.type == "ref"

imported_symbol(imp) := imp.alias

imported_symbol(imp) := regal.last(imp.path.value).value if not imp.alias

to_string(ref) := concat(".", [path |
	some part in ref
	path := part.value
])

resolve(ref, _, _) := to_string(ref) if ref[0].value == "data"

# imported symbol
resolve(ref, _, imported_symbols) := concat(".", resolved) if {
	ref[0].value != "data"

	resolved := array.concat(
		imported_symbols[ref[0].value],
		[part.value | some part in array.slice(ref, 1, count(ref))],
	)
}

# not imported — must be local or package
resolve(ref, pkg_path, imported_symbols) := concat(".", resolved) if {
	ref[0].value != "data"

	not imported_symbols[ref[0].value]

	resolved := array.concat(
		pkg_path,
		[part.value | some part in ref],
	)
}
