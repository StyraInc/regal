# METADATA
# description: Rule named "if"
package regal.rules.bugs["rule-named-if"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	# this is only an optimization — as we already have collected all the rule
	# names, we'll do a fast lookup to know if we need to iterate over the rules
	# at all, which we'll do only to retrieve the location of the rule
	"if" in ast.rule_names

	some rule in input.rules
	ast.name(rule) == "if"

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
