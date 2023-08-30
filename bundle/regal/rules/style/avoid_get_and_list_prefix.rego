# METADATA
# description: Avoid get_ and list_ prefix for rules and functions
package regal.rules.style["avoid-get-and-list-prefix"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules
	strings.any_prefix_match(ast.name(rule), {"get_", "list_"})

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
