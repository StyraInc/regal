# METADATA
# description: Add description of rule here!
# schemas:
# - input: schema.regal.ast
package custom.regal.rules.{{.Category}}{{.Name}}

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	# Or change to imports, packages, comments, etc.
	some rule in input.rules

	# Deny any rule named foo, bar, or baz. This is just an example!
	# Add your own rule logic here.
	ast.ref_to_string(rule.head.ref) in {"foo", "bar", "baz"}

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}
