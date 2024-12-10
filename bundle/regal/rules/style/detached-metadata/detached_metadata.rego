# METADATA
# description: Detached metadata annotation
package regal.rules.style["detached-metadata"]

import data.regal.ast
import data.regal.result
import data.regal.util

report contains violation if {
	some i, block in ast.comments.blocks

	startswith(trim_space(block[0].text), "METADATA")

	last_row := util.to_location_object(regal.last(block).location).row

	# no need to +1 the index here as rows start counting from 1
	trim_space(input.regal.file.lines[last_row]) == ""

	not _allow_detached(last_row, i, ast.comments.blocks, input.regal.file.lines)

	violation := result.fail(rego.metadata.chain(), result.location(block[0]))
}

_annotation_at_row(row) := annotation if {
	some annotation in ast.annotations

	util.to_location_object(annotation.location).row == row
}

# detached metadata is allowed only if another metadata block follows
# directly after the metadata block
_allow_detached(last_row, i, blocks, lines) if {
	next_block := blocks[i + 1]

	startswith(trim_space(next_block[0].text), "METADATA")

	next_block_row := util.to_location_object(next_block[0].location).row
	lines_between := array.slice(lines, last_row, next_block_row - 1)

	every line in lines_between {
		line == ""
	}
}
