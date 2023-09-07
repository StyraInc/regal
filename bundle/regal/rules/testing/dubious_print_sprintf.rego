# METADATA
# description: dubious print sprintf
package regal.rules.testing["dubious-print-sprintf"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules
	some rule_body in rule.body
	rule_body.terms[0].value[0].value == "print"
    rule_body.terms[1].type == "call" 
	rule_body.terms[1].value[0].value[0].value == "sprintf" 

	violation := result.fail(rego.metadata.chain(), result.location(rule_body.terms[1].value[0].value[0]))
}
