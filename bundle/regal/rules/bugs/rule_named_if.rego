# METADATA
# description: Rule named "if"
package regal.rules.bugs["rule-named-if"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

report contains violation if {
	some rule in input.rules
	rule.head.name == "if"

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
