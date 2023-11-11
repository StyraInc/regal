# METADATA
# description: Avoid functions without args
package regal.rules.bugs["zero-arity-function"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	# notably, not ast.functions, as zero-arity functions are treated
	# as regular rules (i.e. they have no `args` key in the head)
	some rule in ast.rules

	text := base64.decode(rule.location.text)

	regex.match(`^[a-zA-z1-9_\.\[\]"]+\(\)`, text)

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}
