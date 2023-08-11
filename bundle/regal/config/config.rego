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

for_rule(metadata) := _with_level(metadata, "ignore") if {
	force_disabled(metadata)
} else := _with_level(metadata, "error") if {
	force_enabled(metadata)
} else := c if {
	m := merged_config.rules[metadata.custom.category][metadata.title]
	c := object.union(m, {"level": rule_level(m)})
} else := {"level": "error"} if {
	# regal ignore:external-reference
	not merged_config.rules[metadata.custom.category][metadata.title]
}

_with_level(metadata, level) := c if {
	m := merged_config.rules[metadata.custom.category][metadata.title]
	c := object.union(m, {"level": level})
} else := {"level": level}

rule_level(cfg) := "error" if {
	not cfg.level
}

rule_level(cfg) := cfg.level

force_disabled(metadata) if {
	metadata.title in data.eval.params.disable
}

force_disabled(metadata) if {
	# regal ignore:external-reference
	data.eval.params.disable_all
	not metadata.custom.category in data.eval.params.enable_category
	not metadata.title in data.eval.params.enable
}

force_disabled(metadata) if {
	metadata.custom.category in data.eval.params.disable_category
	not metadata.title in data.eval.params.enable
}

force_enabled(metadata) if {
	metadata.title in data.eval.params.enable
}

force_enabled(metadata) if {
	# regal ignore:external-reference
	data.eval.params.enable_all
	not metadata.custom.category in data.eval.params.disable_category
	not metadata.title in data.eval.params.disable
}

force_enabled(metadata) if {
	metadata.custom.category in data.eval.params.enable_category
	not metadata.title in data.eval.params.disable
}
