# METADATA
# description: Max rule length exceeded
package regal.rules.style["rule-length"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

cfg := config.for_rule("style", "rule-length")

report contains violation if {
	some rule in input.rules
	lines := split(base64.decode(rule.location.text), "\n")

	line_count(cfg, rule, lines) > cfg["max-rule-length"]

	not generated_body_exception(cfg, rule)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

generated_body_exception(conf, rule) if {
	conf["except-empty-body"] == true
	ast.generated_body(rule)
}

line_count(cfg, _, lines) := count(lines) if cfg["count-comments"] == true

line_count(cfg, rule, lines) := n if {
	not cfg["count-comments"]

	# Note that this assumes } on its own line
	body_start := rule.location.row + 1
	body_end := (body_start + count(lines)) - 3
	body_total := (body_end - body_start) + 1

	# This does not take into account comments that are
	# on the same line as regular code
	body_comments := sum([1 |
		some comment in input.comments

		comment.Location.row >= body_start
		comment.Location.row <= body_end
	])

	n := body_total - body_comments
}
