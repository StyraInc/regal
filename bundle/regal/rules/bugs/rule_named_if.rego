# METADATA
# description: Rule named "if"
package regal.rules.bugs["rule-named-if"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	# this is only an optimization — as we already have collected all the rule
	# names, we'll do a fast lookup to know if we need to iterate over the rules
	# at all, which we'll do only to retrieve the location of the rule
	"if" in ast.rule_names

	some rule in input.rules
	ast.ref_to_string(rule.head.ref) == "if"

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
