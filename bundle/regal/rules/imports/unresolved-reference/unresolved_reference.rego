# METADATA
# description: Unresolved Reference
package regal.rules.imports["unresolved-reference"]

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

# METADATA
# description: collects exported and full of used refs from each module
aggregate contains entry if {
	exported := object.keys(ast.rule_head_locations)

	entry := result.aggregate(rego.metadata.chain(), {
		"exported_rules": exported,
		"expanded_refs": _all_full_path_refs,
		"prefix_tree": {prefix_path |
			some rule_name in exported
			rule_path := split(rule_name, ".")
			some i in numbers.range(1, count(rule_path))
			prefix_path := array.slice(rule_path, 0, i)
		},
	})
}

# an import is shadowed if it shares name with a rule
_shadowed_imports contains rule_name if {
	some rule_name in ast.rule_names
	ast.resolved_imports[rule_name]
}

# an import is shadowed if it shares name with a variable (or function argument)
_shadowed_imports contains var_name if {
	var_name := ast.found.vars[_][_][_].value
	ast.resolved_imports[var_name]
}

_refs contains ref if {
	terms := ast.found.refs[_][_].value
	terms[0].value != "input"

	name := ast.ref_static_to_string(terms)

	not name in ast.builtin_names
	not name in ast.rule_and_function_names
	not terms[0].value in _shadowed_imports

	# util.to_location_row inlined for some extra performance
	row := to_number(regex.replace(terms[0].location, `^(\d+):.*`, "$1"))
	ref := {
		"name": name,
		"text": input.regal.file.lines[row - 1],
		"location": terms[0].location,
	}
}

_all_full_path_refs[ref.name] contains [ref.location, ref.text] if {
	some ref in _refs
	startswith(ref.name, "data.")
}

_all_full_path_refs[expanded_ref] contains [ref.location, ref.text] if {
	some ref in _refs
	path := split(ref.name, ".")
	expanded_ref := concat(".", array.concat(ast.resolved_imports[path[0]], util.rest(path)))
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
aggregate_report contains violation if {
	all_exports := {export | export := input.aggregate[_].aggregate_data.exported_rules[_]}
	prefix_tree := {prefix | prefix := input.aggregate[_].aggregate_data.prefix_tree[_]}

	some entry in input.aggregate
	some name, refs in entry.aggregate_data.expanded_refs

	# ignore everything after the first "[" in the ref name. E.g. foo.bar[0].baz becomes foo.bar
	ref_name := split(name, "[")[0]
	ref_path := split(ref_name, ".")

	# a reference is considered resolved with respect to a rule if
	# 1: it is the prefix of a rule
	# 2: it indexes into a rule - we do not consider the possible data
	# 3: the reference is ignored in the config
	not ref_path in prefix_tree
	not _is_resolved_ref(ref_path, all_exports)
	not _is_excepted(ref_name)

	some ref in refs

	violation := result.fail(rego.metadata.chain(), _to_location_object(ref[0], ref[1], entry.aggregate_source.file))
}

_is_excepted(ref_full_name) if {
	some exception in config.rules.imports["unresolved-reference"]["except-paths"]
	glob.match(exception, [], ref_full_name)
}

_is_resolved_ref(ref_full_path, all_exports) if {
	some i in numbers.range(1, count(ref_full_path))
	path_prefix := concat(".", array.slice(ref_full_path, 0, i))

	path_prefix in all_exports
}

# like util.to_location_object, but with text and file passed in
# as we don't have access to the usual input.regal.file attributes
# in the context of reporting aggregated data
_to_location_object(loc, text, file) := {"location": {
	"file": file,
	"row": row,
	"col": col,
	"text": text,
	"end": {
		"row": end_row,
		"col": end_col,
	},
}} if {
	[r, c, er, ec] := split(loc, ":")

	row := to_number(r)
	col := to_number(c)
	end_row := to_number(er)
	end_col := to_number(ec)
}
