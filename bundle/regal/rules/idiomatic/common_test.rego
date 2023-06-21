package regal.rules.idiomatic.common_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.idiomatic

report(snippet) := report if {
	# regal ignore:external-reference
	report := idiomatic.report with input as ast.policy(snippet)
		with config.for_rule as {"level": "error"}
}
