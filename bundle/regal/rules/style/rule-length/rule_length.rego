# METADATA
# description: Max rule length exceeded
package regal.rules.style["rule-length"]

import data.regal.config
import data.regal.result
import data.regal.util

report contains violation if {
	cfg := config.for_rule("style", "rule-length")

	some rule in input.rules

	rule_location := util.to_location_object(rule.location)
	lines := split(rule_location.text, "\n")

	_line_count(cfg, rule_location.row, lines) > cfg[_max_length_property(rule.head)]

	not _no_body_exception(cfg, rule)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

_no_body_exception(cfg, rule) if {
	cfg["except-empty-body"] == true
	not rule.body
}

default _max_length_property(_) := "max-rule-length"

_max_length_property(head) := "max-test-rule-length" if startswith(head.ref[0].value, "test_")

_line_count(cfg, _, lines) := count(lines) if cfg["count-comments"] == true

_line_count(cfg, rule_row, lines) := n if {
	not cfg["count-comments"]

	# Note that this assumes } on its own line
	body_start := rule_row + 1
	body_end := (body_start + count(lines)) - 3
	body_total := (body_end - body_start) + 1

	# This does not take into account comments that are
	# on the same line as regular code
	body_comments := count([1 |
		some comment in input.comments

		loc := util.to_location_object(comment.location)

		loc.row >= body_start
		loc.row <= body_end
	])

	n := body_total - body_comments
}
