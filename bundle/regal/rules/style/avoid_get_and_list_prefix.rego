# METADATA
# description: Avoid `get_` and `list_` prefix for rules and functions
package regal.rules.style["avoid-get-and-list-prefix"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules
	strings.any_prefix_match(ast.ref_to_string(rule.head.ref), {"get_", "list_"})

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
