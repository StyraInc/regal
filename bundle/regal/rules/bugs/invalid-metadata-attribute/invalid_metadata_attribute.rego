# METADATA
# description: Invalid attribute in metadata annotation
package regal.rules.bugs["invalid-metadata-attribute"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some block in ast.comments.blocks

	startswith(trim_space(block[0].text), "METADATA")

	text := _block_to_string(block)
	attributes := object.keys(yaml.unmarshal(text))

	some attribute in attributes
	not attribute in ast.comments.metadata_attributes

	violation := result.fail(
		rego.metadata.chain(),
		result.location(_find_line(block, attribute)),
	)
}

_block_to_string(block) := concat("\n", [entry.text |
	some i, entry in block
	i > 0
])

_find_line(block, attribute) := [line |
	some line in block
	startswith(trim_space(line.text), sprintf("%s:", [attribute]))
][0]
