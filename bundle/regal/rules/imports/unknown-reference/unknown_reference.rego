# METADATA
# description: Reference to unknown field.
package regal.rules.imports["unknown-reference"]

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

# METADATA
# description: collects exported and full of used refs from each module
aggregate contains content if {
	# regal ignore:unconditional-assignment
	content := result.aggregate(rego.metadata.chain(), {
		"exported_rules": object.keys(ast.rule_head_locations),
		"expanded_refs": _all_full_path_refs,
		"prefix_tree": _prefix_tree,
	})
}

_import_aliases[alias_key] := string_value if {
	some alias_key, value in ast.resolved_imports
	string_value := concat(".", value)
}

_refs contains ref if {
	_ref := ast.found.refs[_][_].value
	_name := ast.ref_static_to_string(_ref)
	not _name in ast.builtin_names
	ref := object.union(result.location(_ref), {
		"name": _name,
		"path": split(_name, "."),
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
	ref_full_path := split(ref_full_name, ".")
	not ref_full_path in prefix_tree
	not _refers_to_rule_content(ref_full_path, all_exports_in_bundle)


	some ref in ref_locations

	violation := result.fail(rego.metadata.chain(), result.location(ref))
}

# METADATA
# description: a reference is valid with respect to a rule if
# 1: it is the prefix of a rule
# 2: it indexes into a rule - we do not consider the possible data
# 3: the reference is ignored in the config
default _is_known_ref(_, _) := false

_is_known_ref(ref_full_name, rule) if {
	startswith(rule, ref_full_name)
}

_is_known_ref(ref_full_name, rule) if {
	startswith(ref_full_name, rule)
}

_is_known_ref(ref_full_name, _) if {
	cfg := config.for_rule("imports", "unknown-reference")

	some exception in cfg["except-imports"]
	glob.match(exception, [], ref_full_name)
}

_prefix_tree contains prefix_path if {
	some rule_name, _ in ast.rule_head_locations
	rule_path := split(rule_name, ".")
	some i in numbers.range(0, count(rule_path))
	prefix_path := array.slice(rule_path, 0, i) 
	# path := rule_path[i]
}

default _refers_to_rule_content(_, _) := false

_refers_to_rule_content(ref_full_path, rules_paths) if {
	some i in numbers.range(0, count(ref_full_path))
	path_prefix := concat(".", array.slice(ref_full_path, 0, i) )

	path_prefix in rules_paths
}