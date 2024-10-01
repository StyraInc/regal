# METADATA
# description: Object literal following `if`
package regal.rules.bugs["if-object-literal"]

import rego.v1

import data.regal.capabilities
import data.regal.result

# METADATA
# description: Missing capability for keyword `if`
# custom:
#   severity: warning
notices contains result.notice(rego.metadata.chain()) if not capabilities.has_if

report contains violation if {
	some rule in input.rules

	count(rule.body) == 1
	rule.body[0].terms.type == "object"

	violation := result.fail(rego.metadata.chain(), result.location(rule.body[0].terms))
}
