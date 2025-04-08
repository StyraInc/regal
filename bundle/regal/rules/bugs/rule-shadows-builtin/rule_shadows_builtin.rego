# METADATA
# description: Rule name shadows built-in
package regal.rules.bugs["rule-shadows-builtin"]

import data.regal.ast
import data.regal.result

report contains violation if {
	head := input.rules[_].head

	ast.ref_to_string(head.ref) in ast.builtin_namespaces

	violation := result.fail(rego.metadata.chain(), result.location(head))
}
