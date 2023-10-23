# METADATA
# description: Multiple tests with same name
package regal.rules.testing["identically-named-tests"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	test_names := [ast.name(rule) | some rule in ast.tests]

	some i, name in test_names

	name in array.slice(test_names, 0, i)

	violation := result.fail(rego.metadata.chain(), result.location(rule_by_name(name, ast.tests)))
}

rule_by_name(name, rules) := regal.last([rule |
	some rule in rules
	rule.head.ref[0].value == name
])
