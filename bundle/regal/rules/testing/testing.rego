package regal.rules.testing

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

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
#   ref: $baseUrl/$category/test-outside-test-package
# custom:
#   category: testing
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	not endswith(package_name, "_test")

	some rule in tests

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}

# METADATA
# title: file-missing-test-suffix
# description: Files containing tests should have a _test.rego suffix
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/file-missing-test-suffix
# custom:
#   category: testing
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	count(tests) > 0

	not endswith(input.regal.file.name, "_test.rego")

	violation := result.fail(rego.metadata.rule(), {"location": {"file": input.regal.file.name}})
}

# METADATA
# title: identically-named-tests
# description: Multiple tests with same name
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/identically-named-tests
# custom:
#   category: testing
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

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
#   ref: $baseUrl/$category/todo-test
# custom:
#   category: testing
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules

	startswith(rule.head.name, "todo_test_")

	# We don't currently have location for rule heads, but this should
	# change soon: https://github.com/open-policy-agent/opa/pull/5811
	violation := result.fail(rego.metadata.rule(), {"location": {"file": input.regal.file.name}})
}

# METADATA
# title: print-or-trace-call
# description: Call to print or trace function
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/print-or-trace-call
# custom:
#   category: testing
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some call in ast.find_builtin_calls(input)

	name := call[0].value[0].value
	name in {"print", "trace"}

	violation := result.fail(rego.metadata.rule(), result.location(call[0].value[0]))
}
