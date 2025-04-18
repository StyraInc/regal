# METADATA
# description: Files containing tests should have a _test.rego suffix
package regal.rules.testing["file-missing-test-suffix"]

import data.regal.ast
import data.regal.config
import data.regal.result

# METADATA
# description: disabled when filename is unknown
# custom:
#   severity: warn
notices contains result.notice(rego.metadata.chain()) if "no_filename" in config.capabilities.special

report contains violation if {
	count(ast.tests) > 0

	not _valid_test_file_name(input.regal.file.name)

	violation := result.fail(rego.metadata.chain(), {"location": {"file": input.regal.file.name}})
}

_valid_test_file_name(filename) if endswith(filename, "_test.rego")
_valid_test_file_name("test.rego") # Styra DAS convention considered OK
