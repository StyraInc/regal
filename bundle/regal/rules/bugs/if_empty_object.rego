# METADATA
# description: Empty object following `if`
package regal.rules.bugs["if-empty-object"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

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
	rule.body[0].terms.value == []

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}
