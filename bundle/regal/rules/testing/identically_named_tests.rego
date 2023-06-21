# METADATA
# description: Multiple tests with same name
package regal.rules.testing["identically-named-tests"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	test_names := [rule.head.name | some rule in ast.tests]

	some i, name in test_names

	name in array.slice(test_names, 0, i)

	# We don't currently have location for rule heads, but this should
	# change soon: https://github.com/open-policy-agent/opa/pull/5811
	violation := result.fail(rego.metadata.chain(), {"location": {"file": input.regal.file.name}})
}
