# METADATA
# description: Files containing tests should have a _test.rego suffix
package regal.rules.testing["file-missing-test-suffix"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	count(ast.tests) > 0

	not _valid_test_file_name(input.regal.file.name)

	violation := result.fail(rego.metadata.chain(), {"location": {"file": input.regal.file.name}})
}

_valid_test_file_name(filename) if endswith(filename, "_test.rego")

# Styra DAS convention considered OK
_valid_test_file_name("test.rego")