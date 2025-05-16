# METADATA
# description: Prefer importing packages over rules
package regal.rules.imports["prefer-package-imports"]

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

# METADATA
# description: collects imports and package paths from each module
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

	entry := result.aggregate(rego.metadata.chain(), {
		"imports": imports_with_location,
		"package_path": ast.package_path,
	})
}

_custom_regal_package_and_import(pkg_path, "regal") if {
	pkg_path[0] == "custom"
	pkg_path[1] == "regal"
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
aggregate_report contains violation if {
	all_package_paths := {entry.aggregate_data.package_path | some entry in input.aggregate}

	some entry in input.aggregate
	some [path, location] in entry.aggregate_data.imports

	_resolves(path, all_package_paths)
	not path in all_package_paths
	not path in _ignored_import_paths

	violation := result.fail(rego.metadata.chain(), {"location": object.union(util.to_location_no_text(location), {
		"file": entry.aggregate_source.file,
		"text": concat("", ["import data.", concat(".", path)]),
	})})
}

# returns true if the path "resolves" to *any* package part of the same length as the path
_resolves(path, pkg_paths) if count([path |
	some pkg_path in pkg_paths
	pkg_path == array.slice(path, 0, count(pkg_path))
]) > 0

_ignored_import_paths contains split(trim_prefix(item, "data."), ".") if {
	some item in config.rules.imports["prefer-package-imports"]["ignore-import-paths"]
}
