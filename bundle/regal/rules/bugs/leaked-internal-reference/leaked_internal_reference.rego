# METADATA
# description: Outside reference to internal rule or function
package regal.rules.bugs["leaked-internal-reference"]

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	_enabled(_valid_test_file_name(input.regal.file.name), _enable_for_test_files)

	value := ast.found.refs[_][_].value

	contains(ast.ref_to_string(value), "._")

	violation := result.fail(rego.metadata.chain(), result.ranged_from_ref(value))
}

report contains violation if {
	_enabled(_valid_test_file_name(input.regal.file.name), _enable_for_test_files)

	value := input.imports[_].path.value

	contains(ast.ref_to_string(value), "._")

	violation := result.fail(rego.metadata.chain(), result.ranged_from_ref(value))
}

default _enabled(_, _) := true

_enabled(true, false) := false

default _valid_test_file_name(_) := false

_valid_test_file_name(filename) if endswith(filename, "_test.rego")
_valid_test_file_name("test.rego") # Styra DAS convention considered OK

default _enable_for_test_files := false

_enable_for_test_files := config.rules.bugs["leaked-internal-reference"]["include-test-files"]
