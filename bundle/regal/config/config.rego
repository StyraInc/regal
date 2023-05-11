package regal.config

import future.keywords.if

default user_config := {}

user_config := data.regal_user_config

merged_config := object.union(data.regal.config.provided, user_config)

for_rule(metadata) := c if {
	m := merged_config.rules[metadata.custom.category][metadata.title]
	c := object.union(m, {"level": rule_level(m)})
}

for_rule(metadata) := {"level": "error"} if {
	# regal ignore:external-reference
	not merged_config.rules[metadata.custom.category][metadata.title]
}

rule_level(cfg) := "error" if {
	not cfg.level
}

rule_level(cfg) := cfg.level
