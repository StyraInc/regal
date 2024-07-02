# METADATA
# description: Annotation without metadata
package regal.rules.bugs["annotation-without-metadata"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some block in ast.comments.blocks

	block[0].Location.col == 1
	ast.comments.annotation_match(trim_space(block[0].Text))

	violation := result.fail(rego.metadata.chain(), result.location(block[0]))
}
