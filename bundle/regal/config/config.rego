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

config_union := object.union(data.regal.config.provided, user_config)

merged_config["ignore"] := config_union.ignore

merged_config["rules"] := merged_rules

# Iterate over the unioned rules and check for any level set to "". This means the user has not provided
# the level for this rule, so we should fall back on the provided level, i.e. that which was
merged_rules[category] := rules if {
	some category, _rules in config_union.rules
	rules := {rule_name: empty_level_to_provided(category, rule_name, rule_conf) |
		some rule_name, rule_conf in _rules
	}
}

empty_level_to_provided(_, _, conf) := conf if conf.level != ""

empty_level_to_provided(category, title, conf) := object.union(
	conf,
	{"level": data.regal.config.provided.rules[category][title].level},
) if {
	conf.level == ""
} else := object.union(conf, {"level": "error"}) if conf.level == ""

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
