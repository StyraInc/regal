# METADATA
# description: Rule assignment conflict
# related_resources:
#   - description: documentation
#     ref: https://www.openpolicyagent.org/projects/regal/rules/bugs/rule-assignment-conflict
package regal.rules.bugs["rule-assignment-conflict"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some target, rules in _rules_by_target
	count(rules) > 1

	_has_assignment_conflict(rules)

	some i, rule in rules
	i > 0

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

_rules_by_target[target] := rules if {
	some rule in input.rules
	not rule.default
	not _dynamic_ref(rule.head.ref)
	_unconditional(rule)

	target := _target(rule)
	rules := [r |
		some r in input.rules
		not r.default
		not _dynamic_ref(r.head.ref)
		_unconditional(r)
		_target(r) == target
	]
}

_target(rule) := sprintf("%s/%d", [ast.ref_to_string(rule.head.ref), count(rule.head.args)]) if rule.head.args

_target(rule) := ast.ref_to_string(rule.head.ref) if not rule.head.args

_dynamic_ref(ref) if array.slice(ref, 1, 100)[_].type in {"call", "var", "ref", "templatestring"}

_has_assignment_conflict(rules) if count(rules) > 1

_unconditional(rule) if {
	not rule.body
	not rule["else"]
	rule.head.value
	_matches_any_arguments(rule)
}

_matches_any_arguments(rule) if not rule.head.args

_matches_any_arguments(rule) if {
	every arg in rule.head.args {
		arg.type == "var"
	}

	count({arg.value | some arg in rule.head.args}) == count(rule.head.args)
}
