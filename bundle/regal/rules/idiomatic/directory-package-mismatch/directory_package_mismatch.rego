# METADATA
# description: Directory structure should mirror package
package regal.rules.idiomatic["directory-package-mismatch"]

import rego.v1

import data.regal.config
import data.regal.result

# METADATA
# description: |
#   emit warning notice when package has more parts than the directory,
#   as this should likely **not** fail
notices contains _notice(message, "warning") if {
	count(_file_path_values) > 0
	count(_file_path_values) < count(_pkg_path_values)

	message := sprintf(
		"package '%s' has more parts than provided directory path '%s'",
		[concat(".", _pkg_path_values), concat("/", _file_path_values)],
	)
}

# METADATA
# description: emit notice when single file is provided, but with no severity
notices contains _notice(message, "none") if {
	count(_file_path_values) == 0

	message := "provided file has no directory components in its path... try linting a directory"
}

report contains violation if {
	# get the last n components from file path, where n == count(_pkg_path_values)
	file_path_length_matched := array.slice(
		_file_path_values,
		count(_file_path_values) - count(_pkg_path_values),
		count(_file_path_values),
	)

	file_path_length_matched != _pkg_path_values

	not _known_file_path_matches(file_path_length_matched, _pkg_path_values)

	violation := result.fail(rego.metadata.chain(), result.location(input["package"].path))
}

_pkg_path := [p.value |
	some i, p in input["package"].path
	i > 0
]

_pkg_path_values := without_test_suffix if {
	cfg := config.for_rule("idiomatic", "directory-package-mismatch")

	cfg["exclude-test-suffix"]

	without_test_suffix := array.concat(
		array.slice(_pkg_path, 0, count(_pkg_path) - 1),
		[trim_suffix(regal.last(_pkg_path), "_test")],
	)
}

_file_path_values := array.slice(parts, 0, count(parts) - 1) if {
	parts := split(input.regal.file.name, input.regal.environment.path_separator)
}

# when a directory path, like `bar/baz`, is shorter than the package
# path, like `foo.bar.baz` this function returns true when the last
# "known" paths match, i.e. in this case `bar/baz` and `bar.baz`
_known_file_path_matches(file_path, pkg_path) if {
	diff := count(pkg_path) - count(file_path)

	diff > 0
	array.slice(pkg_path, diff, count(pkg_path)) == file_path
}

_notice(message, severity) := {
	"category": "idiomatic",
	"description": message,
	"level": "notice",
	"title": "directory-package-mismatch",
	"severity": severity,
}
