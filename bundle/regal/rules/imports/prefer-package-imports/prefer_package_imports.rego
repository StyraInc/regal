# METADATA
# description: Prefer importing packages over rules
package regal.rules.imports["prefer-package-imports"]

import rego.v1

import data.regal.config
import data.regal.result

cfg := config.for_rule("imports", "prefer-package-imports")

aggregate contains entry if {
	package_path := [part.value | some i, part in input["package"].path; i > 0]

	imports_with_location := [imp |
		some _import in input.imports

		_import.path.value[0].value == "data"
		len := count(_import.path.value)
		len > 1
		path := [part.value | some part in array.slice(_import.path.value, 1, len)]

		# Special case for custom rules, where we don't want to flag e.g. `import data.regal.ast`
		# as unknown, even though it's not a package included in evaluation.
		not custom_regal_package_and_import(package_path, path)

		imp := object.union(result.location(_import), {"path": path})
	]

	entry := result.aggregate(rego.metadata.chain(), {
		"imports": imports_with_location,
		"package_path": package_path,
	})
}

custom_regal_package_and_import(pkg_path, path) if {
	pkg_path[0] == "custom"
	pkg_path[1] == "regal"
	path[0] == "regal"
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
aggregate_report contains violation if {
	all_package_paths := {pkg |
		some entry in input.aggregate
		pkg := entry.aggregate_data.package_path
	}

	all_imports := {imp |
		some entry in input.aggregate
		some imp in entry.aggregate_data.imports
	}

	some imp in all_imports
	resolves(imp.path, all_package_paths)
	not imp.path in all_package_paths
	not imp.path in ignored_import_paths

	violation := result.fail(rego.metadata.chain(), {"location": imp.location})
}

# METADATA
# description: |
#   returns true if the path "resolves" to *any*
#   package part of the same length as the path
resolves(path, pkg_paths) if count([path |
	some pkg_path in pkg_paths
	pkg_path == array.slice(path, 0, count(pkg_path))
]) > 0

ignored_import_paths contains path if {
	some item in cfg["ignore-import-paths"]
	path := [part |
		some i, p in split(item, ".")
		part := normalize_part(i, p)
	]
}

normalize_part(0, part) := part if part != "data"

normalize_part(i, part) := part if i > 0
