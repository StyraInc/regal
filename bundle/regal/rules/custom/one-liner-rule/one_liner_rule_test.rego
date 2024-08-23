package regal.rules.custom["one-liner-rule_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.custom["one-liner-rule"] as rule

test_fail_could_be_one_liner if {
	module := ast.policy(`

	import rego.v1

	allow if {
		input.yes
	}
	`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == expected_with_location({"col": 2, "file": "policy.rego", "row": 7, "text": "\tallow if {"})
}

test_fail_could_be_one_liner_all_keywords if {
	module := ast.policy(`

	import rego.v1

	allow if {
		input.yes
	}
	`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == expected_with_location({"col": 2, "file": "policy.rego", "row": 7, "text": "\tallow if {"})
}

test_fail_could_be_one_liner_allman_style if {
	module := ast.policy(`

	import rego.v1

	allow if
	{
		input.yes
	}
	`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == expected_with_location({"col": 2, "file": "policy.rego", "row": 7, "text": "\tallow if"})
}

test_success_if_not_imported if {
	module := ast.policy(`
	allow := true if {
		1 == 1
	}
	`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == set()
}

test_success_too_long_for_a_one_liner if {
	module := ast.with_rego_v1(`
	rule := "quite a long text up here" if {
		some_really_long_rule_name_in_fact_53_characters_long == another_long_rule_but_only_45_characters_long
	}
	`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == set()
}

test_success_too_long_for_a_one_liner_configured_line_length if {
	module := ast.with_rego_v1(`
	rule if {
		some_really_long_rule_name_in_fact_53_characters_long
	}
	`)

	r := rule.report with input as module with config.for_rule as {"level": "error", "max-line-length": 50}
	r == set()
}

test_success_no_one_liner_comment_in_rule_body if {
	module := ast.with_rego_v1(`
	no_one_liner if {
		# Surely one equals one
		1 == 1
	}
	`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == set()
}

test_success_no_one_liner_comment_in_rule_body_same_line if {
	module := ast.with_rego_v1(`
	no_one_liner if {
		1 == 1 # Surely one equals one
	}
	`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == set()
}

test_success_no_one_liner_comment_in_rule_body_line_below if {
	module := ast.with_rego_v1(`
	no_one_liner if {
		1 == 1
		# Surely one equals one
	}
	`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == set()
}

# This will have to be gated with capability version
# later as this will be forced from 1.0
test_success_does_not_use_if if {
	module := ast.policy(`
	allow {
		1 == 1
	}
	`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == set()
}

test_success_already_a_one_liner if {
	module := ast.with_rego_v1(`allow if 1 == 1`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == set()
}

test_has_notice_if_unmet_capability if {
	r := rule.notices with config.capabilities as {}
	r == {{
		"category": "custom",
		"description": "Missing capability for keyword `if`",
		"level": "notice",
		"severity": "warning",
		"title": "one-liner-rule",
	}}
}

expected := {
	"category": "custom",
	"description": "Rule body could be made a one-liner",
	"level": "error",
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/one-liner-rule", "custom"),
	}],
	"title": "one-liner-rule",
}

# regal ignore:external-reference
expected_with_location(location) := {object.union(expected, {"location": location})}
