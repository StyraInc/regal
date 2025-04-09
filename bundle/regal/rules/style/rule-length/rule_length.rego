# METADATA
# description: Max rule length exceeded
package regal.rules.style["rule-length"]

import data.regal.config
import data.regal.result
import data.regal.util

report contains violation if {
	cfg := config.for_rule("style", "rule-length")

	some rule in input.rules

	lines := split(util.to_location_object(rule.location).text, "\n")

	_line_count(cfg, rule, lines) > cfg[_max_length_property(rule.head)]

	not _generated_body_exception(cfg, rule)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

_generated_body_exception(conf, rule) if {
	conf["except-empty-body"] == true
	not rule.body
}

default _max_length_property(_) := "max-rule-length"

_max_length_property(head) := "max-test-rule-length" if startswith(head.ref[0].value, "test_")

_line_count(cfg, _, lines) := count(lines) if cfg["count-comments"] == true

_line_count(cfg, rule, lines) := n if {
	not cfg["count-comments"]

	# Note that this assumes } on its own line
	body_start := util.to_location_object(rule.location).row + 1
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
