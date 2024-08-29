# METADATA
# description: Max rule length exceeded
package regal.rules.style["rule-length"]

import rego.v1

import data.regal.config
import data.regal.result
import data.regal.util

cfg := config.for_rule("style", "rule-length")

report contains violation if {
	some rule in input.rules
	lines := split(base64.decode(util.to_location_object(rule.location).text), "\n")

	line_count(cfg, rule, lines) > cfg["max-rule-length"]

	not generated_body_exception(cfg, rule)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

generated_body_exception(conf, rule) if {
	conf["except-empty-body"] == true
	not rule.body
}

line_count(cfg, _, lines) := count(lines) if cfg["count-comments"] == true

line_count(cfg, rule, lines) := n if {
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
