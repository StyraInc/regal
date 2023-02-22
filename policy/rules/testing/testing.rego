package regal.rules.testing

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal

# Test that rules named test_* does not exist outside of _test packages

package_name := concat(".", [path.value | some path in input["package"].path])

# METADATA
# title: STY-TESTING-001
# description: Test outside of test package
# related_resources:
# - https://docs.styra.com/regal/rules/sty-testing-001
violation contains msg if {
	not endswith(package_name, "_test")

    some rule in input.rules
    startswith(rule.head.name, "test_")

	msg := regal.fail(rego.metadata.rule(), {})
}
