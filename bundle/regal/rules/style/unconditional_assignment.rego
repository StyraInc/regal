# METADATA
# description: Unconditional assignment in rule body
package regal.rules.style["unconditional-assignment"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

report contains violation if {
	some rule in input.rules

	# Single expression in rule body
	# There's going to be a few cases where more expressions
	# are in the body and it still "unconditional", like e.g
	# a `print` call.. but let's keep it simple for now
	count(rule.body) == 1

	# Remove this and consider proper handling of else once
	# https://github.com/open-policy-agent/opa/issues/5777
	# is resolved
	not rule["else"]

	# Var assignment in rule head
	rule.head.value.type == "var"
	rule_head_var := rule.head.value.value

	# If a `with` statement is found in body, back out, as these
	# can't be moved to the rule head
	not rule.body[0]["with"]

	# Which is an assignment (= or :=)
	terms := rule.body[0].terms
	terms[0].type == "ref"
	terms[0].value[0].type == "var"
	terms[0].value[0].value in {"eq", "assign"}

	# Of var declared in rule head
	terms[1].type == "var"
	terms[1].value == rule_head_var

	violation := result.fail(rego.metadata.chain(), result.location(terms[1]))
}
