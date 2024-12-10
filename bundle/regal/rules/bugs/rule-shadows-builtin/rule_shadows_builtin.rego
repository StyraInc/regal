# METADATA
# description: Rule name shadows built-in
package regal.rules.bugs["rule-shadows-builtin"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules

	ast.ref_to_string(rule.head.ref) in ast.builtin_namespaces

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
