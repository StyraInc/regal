package regal.rules.style.common_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style

report(snippet) := report if {
	# regal ignore:external-reference
	report := style.report with input as ast.with_future_keywords(snippet)
		with config.for_rule as {"level": "error", "max-line-length": 80}
}
