# METADATA
# description: Unnecessary wildcard key
package regal.rules.idiomatic["in-wildcard-key"]

import data.regal.ast
import data.regal.result

# some _, v in input
report contains violation if {
	some symbols in ast.found.symbols[_]

	symbol := symbols[0]

	symbol.type == "call"
	symbol.value[0].value[0].type == "var"
	symbol.value[0].value[1].type == "string"
	symbol.value[0].value[0].value == "internal"
	symbol.value[0].value[1].value == "member_3"
	symbol.value[1].type == "var"
	startswith(symbol.value[1].value, "$")

	violation := result.fail(rego.metadata.chain(), result.location(symbol.value[1]))
}
