package regal.rules.testing

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# Test that rules named test_* does not exist outside of _test packages

package_name := concat(".", [path.value | some path in input["package"].path])

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

	some rule in input.rules
	startswith(rule.head.name, "test_")

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}
