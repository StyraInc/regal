# METADATA
# description: Dubious use of print and sprintf
package regal.rules.testing["dubious-print-sprintf"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	# skip traversing refs if no print calls are registered
	"print" in ast.builtin_functions_called

	value := ast.found.refs[_][_]

	value[0].value[0].type == "var"
	value[0].value[0].value == "print"
	value[1].type == "call"
	value[1].value[0].type == "ref"
	value[1].value[0].value[0].type == "var"
	value[1].value[0].value[0].value == "sprintf"

	violation := result.fail(rego.metadata.chain(), result.location(value[1].value[0].value[0]))
}
