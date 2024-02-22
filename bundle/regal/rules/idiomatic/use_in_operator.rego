# METADATA
# description: Use in to check for membership
package regal.rules.idiomatic["use-in-operator"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some terms in eq_exprs_terms

	nl_terms := non_loop_term(terms)
	count(nl_terms) == 1

	nlt := nl_terms[0]
	static_term(nlt.term)

	# Use the non-loop term positon to determine the
	# location of the loop term (3 is the count of terms)
	violation := result.fail(rego.metadata.chain(), result.location(terms[3 - nlt.pos]))
}

eq_exprs_terms contains terms if {
	terms := input.rules[_].body[_].terms

	terms[0].type == "ref"
	terms[0].value[0].type == "var"
	terms[0].value[0].value in {"eq", "equal"}
}

non_loop_term(terms) := [{"pos": i + 1, "term": term} |
	some i, term in array.slice(terms, 1, 3)
	not loop_term(term)
]

loop_term(term) if {
	term.type == "ref"
	term.value[0].type == "var"
	last := regal.last(term.value)
	last.type == "var"
	startswith(last.value, "$")
}

static_term(term) if term.type in {"array", "boolean", "object", "null", "number", "set", "string", "var"}

static_term(term) if {
	term.type == "ref"
	ast.static_ref(term)
}
