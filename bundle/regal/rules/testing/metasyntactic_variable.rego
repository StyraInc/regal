# METADATA
# description: Metasyntactic variable name
package regal.rules.testing["metasyntactic-variable"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

metasyntactic := {
	"foobar",
	"foo",
	"bar",
	"baz",
	"qux",
	"quux",
	"corge",
	"grault",
	"garply",
	"waldo",
	"fred",
	"plugh",
	"xyzzy",
	"thud",
}

report contains violation if {
	some rule in input.rules
	some ref in ast.named_refs(rule.head.ref)

	lower(ref.value) in metasyntactic

	# In case we have chained rule bodies — only flag the location where we have an actual name:
	# foo {
	#    input.x
	# } {
	#    input.y
	# }
	not ast.is_chained_rule_body(rule, input.regal.file.lines)

	violation := result.fail(rego.metadata.chain(), result.location(ref))
}

report contains violation if {
	some rule in input.rules
	some var in ast.find_vars(rule)

	lower(var.value) in metasyntactic

	ast.is_output_var(rule, var, var.location)

	violation := result.fail(rego.metadata.chain(), result.location(var))
}
