# METADATA
# description: Use the `if` keyword
package regal.rules.idiomatic["use-if"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.capabilities
import data.regal.result

# Note: think more about what UX we want when import_rego_v1
# capbility is available. Should we simply just recommend that
# and silence this rule in that case? I'm inclined to say yes.

# METADATA
# description: Missing capability for keyword `if`
# custom:
#   severity: warning
notices contains result.notice(rego.metadata.chain()) if not capabilities.has_if

report contains violation if {
	some rule in input.rules

	not ast.generated_body(rule)

	head_len := count(base64.decode(rule.head.location.text))
	text := trim_space(substring(base64.decode(rule.location.text), head_len, -1))

	not startswith(text, "if")

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}
