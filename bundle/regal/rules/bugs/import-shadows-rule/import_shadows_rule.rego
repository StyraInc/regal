# METADATA
# description: Import shadows rule
package regal.rules.bugs["import-shadows-rule"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some i
	ref := input.rules[i].head.ref

	count(ref) == 1
	ast.ref_to_string(ref) in ast.imported_identifiers

	violation := result.fail(rego.metadata.chain(), result.location(input.rules[i].head))
}
