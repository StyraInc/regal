# METADATA
# description: Prefer importing packages over rules
package regal.rules.imports["prefer-package-imports"]

import data.regal.ast
import data.regal.config
import data.regal.result

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

		imp := object.union(result.location(_import), {"path": path})
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

	all_imports := {imp |
		some entry in input.aggregate
		some imp in entry.aggregate_data.imports
	}

	some imp in all_imports
	_resolves(imp.path, all_package_paths)
	not imp.path in all_package_paths
	not imp.path in _ignored_import_paths

	violation := result.fail(rego.metadata.chain(), {"location": imp.location})
}

# returns true if the path "resolves" to *any* package part of the same length as the path
_resolves(path, pkg_paths) if count([path |
	some pkg_path in pkg_paths
	pkg_path == array.slice(path, 0, count(pkg_path))
]) > 0

_ignored_import_paths contains path if {
	some item in config.rules.imports["prefer-package-imports"]["ignore-import-paths"]
	path := [part |
		some i, p in split(item, ".")
		part := _normalize_part(i, p)
	]
}

_normalize_part(0, part) := part if part != "data"
_normalize_part(i, part) := part if i > 0
