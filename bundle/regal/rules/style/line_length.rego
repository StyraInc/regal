# METADATA
# description: Line too long
package regal.rules.style["line-length"]

import rego.v1

import data.regal.config
import data.regal.result

cfg := config.for_rule("style", "line-length")

default max_line_length := 120

max_line_length := cfg["max-line-length"]

report contains violation if {
	some i, line in input.regal.file.lines

	line != ""

	line_length := count(line)
	line_length > max_line_length

	not has_word_above_threshold(line, cfg)

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

has_word_above_threshold(line, conf) if {
	threshold := conf["non-breakable-word-threshold"]

	some word in split(line, " ")
	count(word) > threshold
}
