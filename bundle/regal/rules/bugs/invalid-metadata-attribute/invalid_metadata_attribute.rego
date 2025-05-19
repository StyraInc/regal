# METADATA
# description: Invalid attribute in metadata annotation
package regal.rules.bugs["invalid-metadata-attribute"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some block in ast.comments.blocks

	startswith(trim_space(block[0].text), "METADATA")

	text := concat("\n", [entry.text |
		some i, entry in block
		i > 0
	])

	some attribute in object.keys(yaml.unmarshal(text))
	not attribute in ast.comments.metadata_attributes

	violation := result.fail(rego.metadata.chain(), result.location([line |
		some line in block
		startswith(trim_space(line.text), concat("", [attribute, ":"]))
	][0]))
}
