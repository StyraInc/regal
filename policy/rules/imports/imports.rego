package regal.rules.imports

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal

future_keywords_wildcard := {"path": {
	"type": "ref",
	"value": [
		{"type": "var", "value": "future"},
		{"type": "string", "value": "keywords"},
	],
}}

# METADATA
# title: STY-IMPORTS-001
# description: Use explicit future keyword imports
# related_resources:
# - https://docs.styra.com/regal/rules/sty-imports-001
violation contains msg if {
	future_keywords_wildcard in input.imports

	msg := regal.fail(rego.metadata.rule(), {})
}
