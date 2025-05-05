# METADATA
# description: Unresolved import
package regal.rules.imports["unresolved-import"]

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

# METADATA
# description: collects imports and exported refs from each module
aggregate contains entry if {
	imports_with_location := [imp |
		some _import in input.imports

		_import.path.value[0].value == "data"
		len := count(_import.path.value)
		len > 1
		path := [part.value | some part in array.slice(_import.path.value, 1, len)]

		# Special case for custom rules, where we don't want to flag e.g. `import data.regal.ast`
		# as unknown, even though it's not a package included in evaluation.
		not _custom_regal_package_and_import(ast.package_path, path[0])

		imp := [path, _import.location]
	]

	exported_refs := {ast.package_path} | {ref |
		some rule in input.rules

		# locations will only contribute to each item in the set being unique,
		# which we don't want here — we only care for distinct ref paths
		some ref in _to_paths(ast.package_path, rule.head.ref)
	}

	entry := result.aggregate(rego.metadata.chain(), {
		"imports": imports_with_location,
		"exported_refs": exported_refs,
	})
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
aggregate_report contains violation if {
	all_known_refs := {path | path := input.aggregate[_].aggregate_data.exported_refs[_]}

	some entry in input.aggregate
	some [path, location] in entry.aggregate_data.imports

	not path in (all_known_refs | _except_imports)

	# cheap operation failed — need to check wildcards here to account
	# for map generating / general ref head rules
	not _wildcard_match(path, all_known_refs, _except_imports)

	violation := result.fail(rego.metadata.chain(), {"location": object.union(util.to_location_no_text(location), {
		"file": entry.aggregate_source.file,
		"text": sprintf("import %s", [concat(".", array.concat(["data"], path))]),
	})})
}

_custom_regal_package_and_import(pkg_path, "regal") if {
	pkg_path[0] == "custom"
	pkg_path[1] == "regal"
}

# the package part will always be included exported refs
# but if we have a rule like foo.bar.baz
# we'll want to include both foo.bar and foo.bar.baz
_to_paths(pkg_path, ref) := util.all_paths(_to_path(pkg_path, ref)) if count(ref) < 3
_to_paths(pkg_path, ref) := [_to_path(pkg_path, p) | some p in util.all_paths(ref)] if count(ref) > 2

_to_path(pkg_path, ref) := array.concat(pkg_path, [str |
	some i, part in ref
	str := _to_string(i, part)
])

_to_string(0, part) := part.value

_to_string(i, part) := part.value if {
	i > 0
	part.type == "string"
}

_to_string(i, part) := "**" if {
	i > 0
	part.type == "var"
}

_except_imports contains exception if {
	some str in config.rules.imports["unresolved-import"]["except-imports"]
	exception := _trim_data(split(str, "."))
}

_trim_data(path) := array.slice(path, 1, count(path)) if path[0] == "data"
_trim_data(path) := path if path[0] != "data"

_wildcard_match(imp_path, all_known_refs, except_imports) if {
	except_imports_wildcards := {path |
		some except in except_imports
		path := concat(".", except)
		contains(path, "*")
	}

	all_known_refs_wildcards := {path |
		some ref in all_known_refs
		path := concat(".", ref)
		contains(path, "*")
	}

	all_wildcard_paths := except_imports_wildcards | all_known_refs_wildcards

	some path in all_wildcard_paths

	# note that we are quite forgiving here, as we'll match the
	# shortest path component containing a wildcard at the end..
	# we may want to make this more strict later, but as this is
	# a new rule with a potentially high impact, let's start like
	# this and then decide if we want to be more strict later, and
	# perhaps offer that as a "strict" option
	glob.match(path, [], concat(".", imp_path))
}
