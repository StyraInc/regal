# METADATA
# description: |
#   base modules for working with Regal's configuration in Rego
#   this includes responsibilities like providing capabilities, or
#   to determine which rules to enable/disable, and what files to
#   ignore
package regal.config

import rego.v1

# METADATA
# description: the rootDir value set on the current linter instance
# scope: document
default root_dir := ""

root_dir := data.internal.root_dir

# METADATA
# description: the base URL for documentation of linter rules
docs["base_url"] := "https://docs.styra.com/regal/rules"

# METADATA
# description: returns the canonical URL for documentation of the given rule
docs["resolve_url"](url, category) := replace(
	replace(url, "$baseUrl", docs.base_url),
	"$category", category,
)

# METADATA
# description: the default configuration with user config merged on top (if provided)
merged_config := data.internal.combined_config

# METADATA
# description: the resolved capabilities sourced from Regal and user configuration
capabilities := object.union(merged_config.capabilities, {"special": _special})

_special contains "no_filename" if input.regal.file.name == "stdin"

default for_rule(_, _) := {"level": "error"}

# METADATA
# description: |
#   Returns the configuration applied (i.e. the provided configuration
#   merged with any user configuration and possibly command line overrides)
#   to the rule matching the category and title.
# scope: document
for_rule(category, title) := _with_level(category, title, "ignore") if {
	_force_disabled(category, title)
} else := _with_level(category, title, "error") if {
	_force_enabled(category, title)
} else := c if {
	# regal ignore:external-reference
	m := merged_config.rules[category][title]
	c := object.union(m, {"level": rule_level(m)})
}

_with_level(category, title, level) := c if {
	# regal ignore:external-reference
	m := merged_config.rules[category][title]
	c := object.union(m, {"level": level})
} else := {"level": level}

# METADATA
# description: returns the level set for rule, otherwise "error"
# scope: document
default rule_level(_) := "error"

rule_level(cfg) := cfg.level

_force_disabled(_, title) if title in data.eval.params.disable # regal ignore:external-reference

_force_disabled(category, title) if {
	# regal ignore:external-reference
	params := data.eval.params

	params.disable_all
	not category in params.enable_category
	not title in params.enable
}

_force_disabled(category, title) if {
	# regal ignore:external-reference
	params := data.eval.params

	category in params.disable_category
	not title in params.enable
}

# regal ignore:external-reference
_force_enabled(_, title) if title in data.eval.params.enable

_force_enabled(category, title) if {
	# regal ignore:external-reference
	params := data.eval.params

	params.enable_all
	not category in params.disable_category
	not title in params.disable
}

_force_enabled(category, title) if {
	# regal ignore:external-reference
	params := data.eval.params

	category in params.enable_category
	not title in params.disable
}
