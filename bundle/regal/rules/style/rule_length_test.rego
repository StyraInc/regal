package regal.rules.style["rule-length_test"]

import future.keywords.if
import future.keywords.in

import data.regal.config

import data.regal.rules.style["rule-length"] as rule

test_fail_rule_longer_than_configured_max_length if {
	module := regal.parse_module("policy.rego", `package p

	my_long_rule {
		# this rule is longer than the configured max length
		# which in this case is only 3 lines
		#
		input.x
	}
	`)

	r := rule.report with input as module with config.for_rule as {
		"level": "error",
		"max-rule-length": 3,
		"count-comments": true,
	}
	r == {{
		"category": "style",
		"description": "Max rule length exceeded",
		"level": "error",
		"location": {"col": 2, "file": "policy.rego", "row": 3, "text": "\tmy_long_rule {"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/rule-length", "style"),
		}],
		"title": "rule-length",
	}}
}

test_success_rule_not_longer_than_configured_max_length if {
	module := regal.parse_module("policy.rego", `package p

	my_short_rule {
		# this rule is not longer than the configured max length
		# which in this case is 30 lines
		#
		input.x
	}
	`)

	r := rule.report with input as module with config.for_rule as {
		"level": "error",
		"max-rule-length": 30,
		"count-comments": true,
	}
	r == set()
}

test_success_rule_longer_than_configured_max_length_but_comments if {
	module := regal.parse_module("policy.rego", `package p

	my_short_rule {
		# this rule is not longer than the configured max length
		# which in this case is 30 lines
		#
		input.x
	}
	`)

	r := rule.report with input as module with config.for_rule as {
		"level": "error",
		"max-rule-length": 2,
		"count-comments": false,
	}
	r == set()
}

test_success_rule_length_equals_max_length if {
	module := regal.parse_module("policy.rego", `package p

	my_tiny_rule := true
	`)

	r := rule.report with input as module with config.for_rule as {
		"level": "error",
		"max-rule-length": 1,
		"count-comments": false,
	}
	r == set()
}
