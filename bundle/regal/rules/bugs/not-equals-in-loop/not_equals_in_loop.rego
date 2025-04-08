# METADATA
# description: Use of != in loop
package regal.rules.bugs["not-equals-in-loop"]

import data.regal.ast
import data.regal.result

report contains violation if {
	terms := ast.exprs[_][_].terms

	terms[0].type == "ref"
	terms[0].value[0].type == "var"
	terms[0].value[0].value == "neq"

	some neq_term in array.slice(terms, 1, count(terms))
	neq_term.type == "ref"

	some value in neq_term.value
	ast.is_wildcard(value)

	violation := result.fail(rego.metadata.chain(), result.location(terms[0]))
}
