# METADATA
# description: Custom rule with type checker failure
# schemas:
# - input: schema.regal.ast
package custom.regal.rules.naming.type_fail

report contains foo if {
	# There is no "foo" attribute in the AST,
	# so this should fail at compile time
	foo := input.foo
	foo == "bar"
}
