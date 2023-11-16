# METADATA
# description: Detached metadata annotation
package regal.rules.style["detached-metadata"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some block in ast.comments.blocks

	startswith(trim_space(block[0].Text), "METADATA")

	last_row := regal.last(block).Location.row

	# no need to +1 the index here as rows start counting from 1
	trim_space(input.regal.file.lines[last_row]) == ""

	annotation := annotation_at_row(block[0].Location.row)
	annotation.scope != "document"

	violation := result.fail(rego.metadata.chain(), result.location(block[0]))
}

annotation_at_row(row) := annotation if {
	some annotation in input.annotations

	annotation.location.row == row
}
