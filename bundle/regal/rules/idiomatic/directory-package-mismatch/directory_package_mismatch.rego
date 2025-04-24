# METADATA
# description: Directory structure should mirror package
package regal.rules.idiomatic["directory-package-mismatch"]

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

# METADATA
# description: disabled when filename is unknown
# custom:
#   severity: warn
notices contains result.notice(rego.metadata.chain()) if "no_filename" in config.capabilities.special

report contains violation if {
	# get the last n components from file path, where n == count(_pkg_path_values)
	file_path_length_matched := array.slice(
		_file_path_values,
		count(_file_path_values) - count(_pkg_path_values),
		count(_file_path_values),
	)

	file_path_length_matched != _pkg_path_values

	violation := result.fail(
		rego.metadata.chain(),
		# skip the "data" part of the path, as it has no location
		result.ranged_from_ref(util.rest(input["package"].path)),
	)
}

_pkg_path_values := ast.package_path if not config.rules.idiomatic["directory-package-mismatch"]["exclude-test-suffix"]

_pkg_path_values := without_test_suffix if {
	config.rules.idiomatic["directory-package-mismatch"]["exclude-test-suffix"] == true

	without_test_suffix := array.concat(
		array.slice(ast.package_path, 0, count(ast.package_path) - 1),
		[trim_suffix(regal.last(ast.package_path), "_test")],
	)
}

_file_path_values := array.slice(parts, 0, count(parts) - 1) if {
	parts := split(input.regal.file.abs, input.regal.environment.path_separator)
}
