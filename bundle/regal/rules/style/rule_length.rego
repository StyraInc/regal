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

	count(lines) > cfg["max-rule-length"]

	not generated_body_exception(cfg, rule)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

generated_body_exception(conf, rule) if {
	conf["except-empty-body"] == true
	ast.generated_body(rule)
}
