# METADATA
# description: Reference to input in test without using with
# related_resources:
#   - description: documentation
#     ref: https://www.openpolicyagent.org/projects/regal/rules/testing/input-without-with
package regal.rules.testing["input-without-with"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some test in ast.tests

	walk(test.body, [_, expr])
	is_array(expr.terms)

	# Expression references input
	some ref in _input_refs(expr.terms)

	# But does not have a `with` targeting input
	not _has_input_with(object.get(expr, "with", []))

	violation := result.fail(rego.metadata.chain(), result.location(ref))
}

_input_refs(terms) := [term |
	walk(terms, [_, term])
	term.type == "ref"
	term.value[0].type == "var"
	term.value[0].value == "input"
]

_has_input_with(with_clauses) if {
	some with_clause in with_clauses
	with_clause.target.type == "ref"
	with_clause.target.value[0].type == "var"
	with_clause.target.value[0].value == "input"
}
