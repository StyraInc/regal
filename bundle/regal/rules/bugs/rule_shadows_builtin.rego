# METADATA
# description: Rule name shadows built-in
package regal.rules.bugs["rule-shadows-builtin"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules
	ast.name(rule) in ast.builtin_namespaces

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
