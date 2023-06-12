package regal.rules.testing

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

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

	test_names := [rule.head.name | some rule in ast.tests]

	some i, name in test_names

	name in array.slice(test_names, 0, i)

	# We don't currently have location for rule heads, but this should
	# change soon: https://github.com/open-policy-agent/opa/pull/5811
	violation := result.fail(rego.metadata.rule(), {"location": {"file": input.regal.file.name}})
}
