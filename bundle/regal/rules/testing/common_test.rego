package regal.rules.testing.common_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.testing

report(snippet) := report if {
	# regal ignore:external-reference
	report := testing.report with input as ast.with_future_keywords(snippet)
		with config.for_rule as {"level": "error"}
}
