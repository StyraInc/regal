# METADATA
# description: Use the `if` keyword
package regal.rules.idiomatic["use-if"]

import rego.v1

import data.regal.ast
import data.regal.capabilities
import data.regal.result
import data.regal.util

# METADATA
# description: Missing capability for keyword `if`
# custom:
#   severity: warning
notices contains result.notice(rego.metadata.chain()) if not capabilities.has_if

report contains violation if {
	# if rego.v1 is imported, OPA will ensure this anyway
	not ast.imports_has_path(ast.imports, ["rego", "v1"])

	some rule in input.rules
	rule.body

	head_len := count(base64.decode(util.to_location_object(rule.head.location).text))
	text := trim_space(substring(base64.decode(util.to_location_object(rule.location).text), head_len, -1))

	not startswith(text, "if")

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}
