package regal.config

import future.keywords.if
import future.keywords.in

default user_config := {}

docs["base_url"] := "https://docs.styra.com/regal/rules"

docs["resolve_url"](url, category) := replace(
	replace(url, "$baseUrl", docs.base_url),
	"$category", category,
)

user_config := data.regal_user_config

merged_config := object.union(data.regal.config.provided, user_config)

# METADATA
# description: |
#   Returns the configuration applied (i.e. the provided configuration
#   merged with any user configuration and possibly command line overrides)
#   to the rule matching the category and title.
for_rule(category, title) := _with_level(category, title, "ignore") if {
	force_disabled(category, title)
} else := _with_level(category, title, "error") if {
	force_enabled(category, title)
} else := c if {
	m := merged_config.rules[category][title]
	c := object.union(m, {"level": rule_level(m)})
} else := {"level": "error"} if {
	# regal ignore:external-reference
	not merged_config.rules[category][title]
}

_with_level(category, title, level) := c if {
	m := merged_config.rules[category][title]
	c := object.union(m, {"level": level})
} else := {"level": level}

rule_level(cfg) := "error" if {
	not cfg.level
}

rule_level(cfg) := cfg.level

force_disabled(_, title) if title in data.eval.params.disable

force_disabled(category, title) if {
	# regal ignore:external-reference
	data.eval.params.disable_all
	not category in data.eval.params.enable_category
	not title in data.eval.params.enable
}

force_disabled(category, title) if {
	category in data.eval.params.disable_category
	not title in data.eval.params.enable
}

force_enabled(_, title) if title in data.eval.params.enable

force_enabled(category, title) if {
	# regal ignore:external-reference
	data.eval.params.enable_all
	not category in data.eval.params.disable_category
	not title in data.eval.params.disable
}

force_enabled(category, title) if {
	category in data.eval.params.enable_category
	not title in data.eval.params.disable
}
