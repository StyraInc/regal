# METADATA
# scope: subpackages
# authors:
# - Styra
# related_resources:
# - https://www.styra.com
package regal

fail(metadata, details) := violation {
	with_location := object.union(metadata, details)
	with_category := object.union(with_location, {"category": with_location.custom.category})

	violation := object.remove(with_category, ["custom", "scope"])
}

ast(policy) := rego.parse_module("policy.rego", concat("", [
	`package policy

	import future.keywords.contains
	import future.keywords.every
	import future.keywords.if
	import future.keywords.in

	`,
	policy,
]))

is_snake_case(str) := str == lower(str)

rule_config(metadata) := data.regal.config.rules[metadata.custom.category][metadata.title]
