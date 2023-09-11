# METADATA
# description: Avoid chaining rule bodies
package regal.rules.style["chained-rule-body"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules

	ast.is_chained_rule_body(rule, input.regal.file.lines)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
