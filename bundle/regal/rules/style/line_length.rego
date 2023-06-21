package regal.rules.style

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# METADATA
# title: line-length
# description: Line too long
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/line-length
# custom:
#   category: style
report contains violation if {
	cfg := config.for_rule(rego.metadata.rule())

	cfg.level != "ignore"

	some i, line in input.regal.file.lines

	line_length := count(line)
	line_length > cfg["max-line-length"]

	violation := result.fail(
		rego.metadata.rule(),
		{"location": {
			"file": input.regal.file.name,
			"row": i + 1,
			"col": line_length,
			"text": input.regal.file.lines[i],
		}},
	)
}
