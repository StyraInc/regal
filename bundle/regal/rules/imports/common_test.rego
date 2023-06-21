package regal.rules.imports.common_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.imports

report_with_fk(snippet) := report if {
	# regal ignore:external-reference
	report := imports.report with input as ast.with_future_keywords(snippet) with config.for_rule as {"level": "error"}
}

report(snippet) := report if {
	# regal ignore:external-reference
	report := imports.report with input as ast.policy(snippet) with config.for_rule as {"level": "error"}
}
