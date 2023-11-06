# METADATA
# description: Unnecessary use of `some`
package regal.rules.style["unnecessary-some"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	# No need to traverse rules here if we're not importing `in`
	ast.imports_keyword(input.imports, "in")

	some rule in input.rules

	walk(rule, [_, value])

	value[0].type == "call"
	value[0].value[0].type == "ref"

	some_is_unnecessary(value)

	violation := result.fail(rego.metadata.chain(), result.location(value))
}

some_is_unnecessary(value) if {
	ref := value[0].value[0].value

	[ref[0].value, ref[1].value] == ["internal", "member_2"]

	value[0].value[1].type in ast.scalar_types
}

some_is_unnecessary(value) if {
	ref := value[0].value[0].value

	[ref[0].value, ref[1].value] == ["internal", "member_3"]

	value[0].value[1].type in ast.scalar_types
	value[0].value[2].type in ast.scalar_types
}
