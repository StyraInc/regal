# METADATA
# description: Entrypoint can't be marked internal
package regal.rules.bugs["internal-entrypoint"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in ast.rules
	some annotation in rule.annotations

	annotation.entrypoint == true
	startswith(ast.ref_to_string(rule.head.ref), "_")

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
