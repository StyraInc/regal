# METADATA
# description: Test outside of test package
package regal.rules.testing["test-outside-test-package"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	not endswith(ast.package_name, "_test")

	some rule in ast.tests

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
