# METADATA
# description: |
#   the `regov1`` provider provides completion suggestions for
#   `rego.v1` following an `import` declaration
package regal.lsp.completion.providers.regov1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

# METADATA
# description: completion suggestion for rego.v1
items contains item if {
	not strings.any_prefix_match(input.regal.file.lines, "import rego.v1")
	not input.regal.context.rego_version == 3 # the rego.v1 import is not used in v1 rego

	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	startswith(line, "import ")

	word := location.ref_at(line, input.regal.context.location.col)

	startswith("rego.v1", word.text)

	item := {
		"label": "rego.v1",
		"kind": kind.module,
		"detail": "use rego.v1",
		"textEdit": {
			"range": location.word_range(word, position),
			"newText": "rego.v1\n\n",
		},
	}
}
