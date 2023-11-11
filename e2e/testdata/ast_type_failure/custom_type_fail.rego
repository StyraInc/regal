# METADATA
# description: Custom rule with type checker failure
# schemas:
# - input: schema.regal.ast
package custom.regal.rules.naming.type_fail

import rego.v1

report contains foo if {
	# There is no "foo" attrinbute in the AST,
	# so this should fail at compile time
	foo := input.foo
	foo == "bar"
}
