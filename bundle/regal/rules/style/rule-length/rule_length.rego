# METADATA
# description: Max rule length exceeded
package regal.rules.style["rule-length"]

import data.regal.config
import data.regal.result
import data.regal.util

report contains violation if {
	cfg := config.rules.style["rule-length"]

	some rule in input.rules

	text := util.to_location_object(rule.location).text

	_line_count(cfg, text) > cfg[_max_length_property(rule.head.ref[0].value)]

	not _no_body_exception(cfg, rule)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

_no_body_exception(cfg, rule) if {
	cfg["except-empty-body"] == true
	not rule.body
}

default _max_length_property(_) := "max-rule-length"

_max_length_property(value) := "max-test-rule-length" if startswith(value, "test_")

_line_count(cfg, text) := strings.count(text, "\n") + 1 if cfg["count-comments"] == true

_line_count(cfg, text) := n if {
	not cfg["count-comments"]

	n := count([1 |
		some line in split(text, "\n")
		not startswith(trim_space(line), "#")
	])
}
