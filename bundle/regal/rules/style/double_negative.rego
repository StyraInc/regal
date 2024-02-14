# METADATA
# description: Avoid double negatives
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/community/double-negative
# schemas:
# - input: schema.regal.ast
package regal.rules.style["double-negative"]

import rego.v1

import data.regal.result

report contains violation if {
	walk(input.rules, [_, node])

	node.negated

	node.terms.type == "var"
	strings.any_prefix_match(node.terms.value, negatives)

	violation := result.fail(rego.metadata.chain(), result.location(node))
}

negatives := {
	"cannot_",
	"no_",
	"non_",
	"not_",
}
