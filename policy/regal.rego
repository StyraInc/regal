# METADATA
# scope: subpackages
# authors:
# - Styra
# related_resources:
# - https://www.styra.com
package regal

fail(metadata, details) := object.union(object.remove(metadata, ["scope"]), details)

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
