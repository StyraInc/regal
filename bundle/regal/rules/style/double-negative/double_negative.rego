# METADATA
# description: Avoid double negatives
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/style/double-negative
# schemas:
# - input: schema.regal.ast
package regal.rules.style["double-negative"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some node
	ast.negated_expressions[_][node]

	node.terms.type == "var"
	strings.any_prefix_match(node.terms.value, {
		"cannot_",
		"no_",
		"non_",
		"not_",
	})

	violation := result.fail(rego.metadata.chain(), result.location(node))
}
