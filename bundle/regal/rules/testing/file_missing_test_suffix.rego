package regal.rules.testing

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

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

	count(ast.tests) > 0

	not endswith(input.regal.file.name, "_test.rego")

	violation := result.fail(rego.metadata.rule(), {"location": {"file": input.regal.file.name}})
}
