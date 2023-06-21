package regal.rules.testing

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

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

	not endswith(ast.package_name, "_test")

	some rule in ast.tests

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}
