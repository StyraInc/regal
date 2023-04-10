package regal.rules.testing

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# Test that rules named test_* does not exist outside of _test packages

package_name := concat(".", [path.value | some path in input["package"].path])

tests := [rule |
	some rule in input.rules
	startswith(rule.head.name, "test_")
]

# METADATA
# title: test-outside-test-package
# description: Test outside of test package
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/test-outside-test-package
# custom:
#   category: testing
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	not endswith(package_name, "_test")

	some rule in tests

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}

# METADATA
# title: identically-named-tests
# description: Multiple tests with same name
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/identically-named-tests
# custom:
#   category: testing
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	test_names := [rule.head.name | some rule in tests]

	some i, name in test_names

	name in array.slice(test_names, 0, i)

	# We don't currently have location for rule heads, but this should
	# change soon: https://github.com/open-policy-agent/opa/pull/5811
	violation := result.fail(rego.metadata.rule(), {"location": {"file": input.regal.file.name}})
}

# METADATA
# title: todo-test
# description: TODO test encountered
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/todo-test
# custom:
#   category: testing
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules

	startswith(rule.head.name, "todo_test_")

	# We don't currently have location for rule heads, but this should
	# change soon: https://github.com/open-policy-agent/opa/pull/5811
	violation := result.fail(rego.metadata.rule(), {"location": {"file": input.regal.file.name}})
}
