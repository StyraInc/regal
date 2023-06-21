package regal.rules.testing

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

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

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}
