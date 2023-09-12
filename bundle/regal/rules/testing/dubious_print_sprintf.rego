# METADATA
# description: Dubious use of print and sprintf
package regal.rules.testing["dubious-print-sprintf"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	walk(input.rules, [_, value])

	value[0].type == "ref"
	value[0].value[0].type == "var"
	value[0].value[0].value == "print"
	value[1].type == "call"
	value[1].value[0].type == "ref"
	value[1].value[0].value[0].type == "var"
	value[1].value[0].value[0].value == "sprintf"

	violation := result.fail(rego.metadata.chain(), result.location(value[1].value[0].value[0]))
}
