# METADATA
# description: Empty object following `if`
package regal.rules.bugs["if-empty-object"]

import rego.v1

import data.regal.capabilities
import data.regal.result

# METADATA
# description: Missing capability for keyword `if`
# custom:
#   severity: warning
notices contains result.notice(rego.metadata.chain()) if not capabilities.has_if

# METADATA
# description: |
#   NOTE: this rule has been deprecated and is no longer enabled by default
#   Use the `if-object-literal` rule instead, which checks for any object,
#   non-empty or not
report contains violation if {
	some rule in input.rules

	count(rule.body) == 1

	rule.body[0].terms.type == "object"
	rule.body[0].terms.value == []

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}
