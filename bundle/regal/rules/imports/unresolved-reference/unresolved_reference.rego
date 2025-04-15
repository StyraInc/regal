# METADATA
# description: Unresolved Reference
package regal.rules.imports["unresolved-reference"]

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

# METADATA
# description: collects exported and full of used refs from each module
aggregate contains result.aggregate(rego.metadata.chain(), {
	"exported_rules": object.keys(ast.rule_head_locations),
	"expanded_refs": _all_full_path_refs,
	"prefix_tree": _prefix_tree,
})

_import_aliases[alias_key] := string_value if {
	some alias_key, value in ast.resolved_imports
	string_value := concat(".", value)
}

# an import is shadowed if it shares name with a rule
_shadowed_imports contains rule_name if {
	some rule_name in ast.rule_names
	_import_aliases[rule_name]
}

# an import is shadowed if it shares name with a variable (or function argument)
_shadowed_imports contains var_name if {
	var_name := ast.found.vars[_][_][_].value
	_import_aliases[var_name]
}

_refs contains ref if {
	term := ast.found.refs[_][_].value
	name := ast.ref_static_to_string(term)

	not name in ast.builtin_names

	path := split(name, ".")
	not path[0] in _shadowed_imports

	ref := object.union(result.location(term), {
		"name": name,
		"path": path,
	})
}

_all_full_path_refs[ref.name] contains ref if {
	some ref in _refs
	ref.path[0] == "data"
}

_all_full_path_refs[expanded_ref] contains ref if {
	some ref in _refs
	full_source_prefix := _import_aliases[ref.path[0]]

	full_path_array := array.concat([full_source_prefix], util.rest(ref.path))

	expanded_ref := concat(".", full_path_array)
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
aggregate_report contains violation if {
	all_exports_in_bundle := {export |
		some entry in input.aggregate
		some export in entry.aggregate_data.exported_rules
	}

	prefix_tree := {prefix |
		some entry in input.aggregate
		some prefix in entry.aggregate_data.prefix_tree
	}

	some entry in input.aggregate
	some ref_full_name, ref_locations in entry.aggregate_data.expanded_refs

	# Ignore everything after the first "[" in the ref name. E.g. foo.bar[0].baz becomes foo.bar
	simplified_ref_name := split(ref_full_name, "[")[0]

	not _is_resolved_ref(simplified_ref_name, prefix_tree, all_exports_in_bundle)

	some ref in ref_locations
	violation := result.fail(rego.metadata.chain(), result.location(ref))
}

# METADATA
# description: a reference is valid with respect to a rule if
# custom:
#   1: it is the prefix of a rule
#   2: it indexes into a rule - we do not consider the possible data
#   3: the reference is ignored in the config
_is_resolved_ref(ref_full_name, _, _) if {
	some exception in config.rules.imports["unresolved-reference"]["except-paths"]
	glob.match(exception, [], ref_full_name)
}

_is_resolved_ref(ref_full_name, prefix_tree, _) if {
	ref_full_path := split(ref_full_name, ".")
	ref_full_path in prefix_tree
}

_is_resolved_ref(ref_full_name, _, all_exports_in_bundle) if {
	ref_full_path := split(ref_full_name, ".")

	some i in numbers.range(0, count(ref_full_path))
	path_prefix := concat(".", array.slice(ref_full_path, 0, i))

	path_prefix in all_exports_in_bundle
}

_prefix_tree contains prefix_path if {
	some rule_name, _ in ast.rule_head_locations
	rule_path := split(rule_name, ".")
	some i in numbers.range(0, count(rule_path))
	prefix_path := array.slice(rule_path, 0, i)
}
