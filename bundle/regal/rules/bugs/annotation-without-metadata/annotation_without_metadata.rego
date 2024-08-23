# METADATA
# description: Annotation without metadata
package regal.rules.bugs["annotation-without-metadata"]

import rego.v1

import data.regal.ast
import data.regal.result
import data.regal.util

report contains violation if {
	some block in ast.comments.blocks

	location := util.to_location_object(block[0].location)

	location.col == 1
	ast.comments.annotation_match(trim_space(block[0].text))

	violation := result.fail(rego.metadata.chain(), result.location(block[0]))
}
