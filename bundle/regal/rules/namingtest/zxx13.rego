# METADATA
# description: Add description of rule here!
package regal.rules.namingtest.zxx13

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	# Or change to imports, packages, comments, etc.
	some rule in input.rules

	# Deny any rule named foo, bar, or baz. This is just an example!
	# Add your own rule logic here.
	ast.name(rule) in {"foo", "bar", "baz"}

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}
