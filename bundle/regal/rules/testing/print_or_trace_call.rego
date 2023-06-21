package regal.rules.testing

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

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
