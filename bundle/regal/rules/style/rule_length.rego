# METADATA
# description: Max rule length exceeded
package regal.rules.style["rule-length"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

cfg := config.for_rule("style", "rule-length")

report contains violation if {
	some rule in input.rules
	lines := split(base64.decode(rule.location.text), "\n")

	count(lines) > cfg["max-rule-length"]

	not empty_body_exception(cfg, rule)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

empty_body_exception(conf, rule) if {
	conf["except-empty-body"] == true
	count(rule.body) == 1
	rule.body[0].terms.type == "boolean"
	rule.body[0].terms.value == true
}
