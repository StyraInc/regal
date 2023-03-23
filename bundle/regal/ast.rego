package regal.ast

with_future_keywords(policy) := regal.parse_module("policy.rego", concat("", [
	`package policy

import future.keywords.contains
import future.keywords.every
import future.keywords.if
import future.keywords.in

`,
	policy,
]))
