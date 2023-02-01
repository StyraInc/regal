# METADATA
# scope: subpackages
# authors:
# - Styra
# related_resources:
# - https://www.styra.com
package regal

fail(metadata, details) := object.union(metadata, details)

ast(policy) := rego.parse_module("policy.rego", concat("", [
	"package policy\n\n",
	policy,
]))

is_snake_case(str) := str == lower(str)