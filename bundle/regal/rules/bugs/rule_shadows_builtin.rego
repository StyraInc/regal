# METADATA
# description: Rule name shadows built-in
package regal.rules.bugs["rule-shadows-builtin"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules
	rule.head.name in ast.builtin_names

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
