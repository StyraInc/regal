# METADATA
# description: Use the `contains` keyword
package regal.rules.idiomatic["use-contains"]

import data.regal.ast
import data.regal.capabilities
import data.regal.result
import data.regal.util

# METADATA
# description: Missing capability for keyword `contains`
# custom:
#   severity: warning
notices contains result.notice(rego.metadata.chain()) if not capabilities.has_contains

# METADATA
# description: Rule made obsolete by OPA 1.0
# custom:
#   severity: none
notices contains result.notice(rego.metadata.chain()) if capabilities.is_opa_v1

report contains violation if {
	# if rego.v1 is imported, OPA will ensure this anyway
	not ast.imports_has_path(ast.imports, ["rego", "v1"])

	some rule in ast.rules

	rule.head.key
	not rule.head.value

	text := split(util.to_location_object(rule.location).text, "\n")[0]

	not contains(text, " contains ")

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
