# METADATA
# description: Max rule length exceeded
package regal.rules.style["rule-length"]

import rego.v1

import data.regal.config
import data.regal.result
import data.regal.util

report contains violation if {
	cfg := config.for_rule("style", "rule-length")

	some rule in input.rules
	lines := split(base64.decode(util.to_location_object(rule.location).text), "\n")

	_line_count(cfg, rule, lines) > cfg["max-rule-length"]

	not _generated_body_exception(cfg, rule)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

_generated_body_exception(conf, rule) if {
	conf["except-empty-body"] == true
	not rule.body
}

_line_count(cfg, _, lines) := count(lines) if cfg["count-comments"] == true

_line_count(cfg, rule, lines) := n if {
	not cfg["count-comments"]

	# Note that this assumes } on its own line
	body_start := util.to_location_object(rule.location).row + 1
	body_end := (body_start + count(lines)) - 3
	body_total := (body_end - body_start) + 1

	# This does not take into account comments that are
	# on the same line as regular code
	body_comments := sum([1 |
		some comment in input.comments

		loc := util.to_location_object(comment.location)

		loc.row >= body_start
		loc.row <= body_end
	])

	n := body_total - body_comments
}
