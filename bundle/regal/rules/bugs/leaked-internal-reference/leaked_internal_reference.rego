# METADATA
# description: Outside reference to internal rule or function
package regal.rules.bugs["leaked-internal-reference"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	_enabled(_valid_test_file_name(input.regal.file.name), _enable_for_test_files)

	ref := ast.found.refs[_][_]

	contains(ast.ref_to_string(ref.value), "._")

	violation := result.fail(rego.metadata.chain(), result.ranged_from_ref(ref.value))
}

report contains violation if {
	_enabled(_valid_test_file_name(input.regal.file.name), _enable_for_test_files)

	some imported in input.imports

	contains(ast.ref_to_string(imported.path.value), "._")

	violation := result.fail(rego.metadata.chain(), result.ranged_from_ref(imported.path.value))
}

default _enabled(_, _) := true

_enabled(true, false) := false

default _valid_test_file_name(_) := false

_valid_test_file_name(filename) if endswith(filename, "_test.rego")

# Styra DAS convention considered OK
_valid_test_file_name("test.rego")

_cfg := config.for_rule("bugs", "leaked-internal-reference")

default _enable_for_test_files := false

_enable_for_test_files := object.get(_cfg, "include-test-files", false)
