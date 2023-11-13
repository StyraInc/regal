# METADATA
# description: Use the `contains` keyword
package regal.rules.idiomatic["use-contains"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.capabilities
import data.regal.result

# METADATA
# description: Missing capability for keyword `contains`
# custom:
#   severity: warning
notices contains result.notice(rego.metadata.chain()) if not capabilities.has_contains

report contains violation if {
	some rule in ast.rules

	rule.head.key
	not rule.head.value

	text := base64.decode(rule.head.location.text)

	not contains(text, " contains ")

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}
