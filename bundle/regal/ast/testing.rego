package regal.ast

import rego.v1

# METADATA
# description: parses provided policy with all future keywords imported. Primarily for testing.
with_rego_v1(policy) := regal.parse_module("policy.rego", concat("", [
	`package policy

import rego.v1

`,
	policy,
]))

# METADATA
# description: parse provided snippet with a generic package declaration added
policy(snippet) := regal.parse_module("policy.rego", concat("", [
	"package policy\n\n",
	snippet,
]))
