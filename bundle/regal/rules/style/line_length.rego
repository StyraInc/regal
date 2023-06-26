# METADATA
# description: Line too long
package regal.rules.style["line-length"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

report contains violation if {
	cfg := config.for_rule({"category": "style", "title": "line-length"})

	some i, line in input.regal.file.lines

	line_length := count(line)
	line_length > cfg["max-line-length"]

	violation := result.fail(
		rego.metadata.chain(),
		{"location": {
			"file": input.regal.file.name,
			"row": i + 1,
			"col": line_length,
			"text": input.regal.file.lines[i],
		}},
	)
}
