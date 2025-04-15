# METADATA
# description: Invalid reference to import.
package regal.rules.imports["unknown-reference"]

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

aggregate contains content if {
	exported_rules := object.keys(ast.rule_head_locations)

	refs := {ref |
		_ref := ast.found.refs[_][_].value
		_name := ast.ref_static_to_string(_ref)
		not _name in ast.builtin_names
		ref := {
			"name": _name,
			"path": split(_name, "."),
			"location": _ref[0].location,
		}
	}

	import_aliases := {alias_key: string_value |
		some alias_key, value in ast.resolved_imports
		string_value := concat(".", value)
	}

	data_refs := {{"full_name": ref.name, "ref": ref} |
		some ref in refs
		ref.path[0] == "data"
	}

	aliased_refs := {{"full_name": expanded_ref, "ref": ref} |
		some ref in refs
		full_source_prefix := import_aliases[ref.path[0]]

		full_path_array := array.concat([full_source_prefix], util.rest(ref.path))

		expanded_ref := concat(".", full_path_array)
	}

	# Combine 'aliased_refs' with 'data_refs'.
	# Make it a map from full path to locations this endpoint is used for performance
	all_refs_full_path := {ref.full_name |
		some ref in (data_refs | aliased_refs)
	}

	all_full_path_refs := {full_name: _refs |
		some full_name in all_refs_full_path
		_refs := {ref.ref |
			some ref in (data_refs | aliased_refs)
			ref.full_name == full_name
		}
	}

	content := result.aggregate(rego.metadata.chain(), {
		"exported_rules": exported_rules,
		"expanded_refs": all_full_path_refs,
	})
}

aggregate_report contains violation if {
	all_exports_in_bundle := {export |
		some entry in input.aggregate
		some export in entry.aggregate_data.exported_rules
	}

	some entry in input.aggregate
	some ref_full_name, ref_locations in entry.aggregate_data.expanded_refs

	# TODO: is pre-fix tree a posibility here?
	# rough calculation of running it a few times in terminal, this loop is almost 50% of the runtime
	every endpoint in all_exports_in_bundle {
		not is_valid_ref(ref_full_name, endpoint)
	}

	some ref in ref_locations

	# todo: this should probably be done with result.location but it doesnt work
	loc := {"location": {
		"file": sprintf("%s:%s", [entry.aggregate_source.file, ref.location]),
		"text": sprintf("%s (%s) does not exist", [ref.name, ref_full_name]),
	}}

	violation := result.fail(rego.metadata.chain(), loc)
}

# METADATA
# description: a reference is valid with respect to a rule if
# 1: it is the prefix of a rule
# 2: it indexes into a rule - we do not consider the possible data
# 3: the reference is ignored in the config
default is_valid_ref(_, _) := false

is_valid_ref(ref_full_name, rule) if {
	startswith(rule, ref_full_name)
}

is_valid_ref(ref_full_name, rule) if {
	startswith(ref_full_name, rule)
}

is_valid_ref(ref_full_name, _) if {
	cfg := config.for_rule("imports", "unknown-ref")

	some exception in cfg["except-imports"]
	glob.match(exception, [], ref_full_name)
}
