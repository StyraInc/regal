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
# title: implicit-future-keywords
# description: Use explicit future keyword imports
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/implicit-future-keywords
# custom:
#   category: imports
report contains violation if {
	regal.rule_config(rego.metadata.rule()).enabled == true

	future_keywords_wildcard in input.imports

	violation := regal.fail(rego.metadata.rule(), {})
}

# METADATA
# title: avoid-importing-input
# description: Avoid importing input
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/avoid-importing-input
# custom:
#   category: imports
report contains violation if {
	regal.rule_config(rego.metadata.rule()).enabled == true

	some imported in input.imports

	imported.path.value[0].value == "input"

	# If we want to allow aliasing input, eg `import input as tfplan`:
	# count(imported.path.value) == 1
    # imported.alias

	violation := regal.fail(rego.metadata.rule(), {})
}
