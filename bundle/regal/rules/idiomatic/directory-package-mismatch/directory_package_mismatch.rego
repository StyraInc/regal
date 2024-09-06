# METADATA
# description: Directory structure should mirror package
package regal.rules.idiomatic["directory-package-mismatch"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	# get the last n components from file path, where n == count(_pkg_path_values)
	file_path_length_matched := array.slice(
		_file_path_values,
		count(_file_path_values) - count(_pkg_path_values),
		count(_file_path_values),
	)

	file_path_length_matched != _pkg_path_values

	violation := result.fail(rego.metadata.chain(), result.location(input["package"].path))
}

_pkg_path_values := ast.package_path if {
	not config.for_rule("idiomatic", "directory-package-mismatch")["exclude-test-suffix"]
}

_pkg_path_values := without_test_suffix if {
	config.for_rule("idiomatic", "directory-package-mismatch")["exclude-test-suffix"]

	without_test_suffix := array.concat(
		array.slice(ast.package_path, 0, count(ast.package_path) - 1),
		[trim_suffix(regal.last(ast.package_path), "_test")],
	)
}

_file_path_values := array.slice(parts, 0, count(parts) - 1) if {
	parts := split(input.regal.file.abs, input.regal.environment.path_separator)
}
