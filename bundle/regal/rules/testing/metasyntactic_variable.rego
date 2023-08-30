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

	violation := result.fail(rego.metadata.chain(), result.location(location_of(ref, rule)))
}

report contains violation if {
	some rule in input.rules
	some var in ast.find_vars(rule)

	lower(var.value) in metasyntactic

	ast.is_output_var(rule, var, var.location)

	violation := result.fail(rego.metadata.chain(), result.location(var))
}

# annoyingly, rule head refs are missing location when a rule.head.name is present,
# or rather when there's only a single item in the ref.. this inconsistency should
# probably be fixed in OPA, but until then, there's this.
location_of(ref, rule) := ref if ref.location

else := rule.head
