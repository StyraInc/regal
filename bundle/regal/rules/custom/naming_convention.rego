# METADATA
# description: Naming convention violation
package regal.rules.custom["naming-convention"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result

cfg := config.for_rule("custom", "naming-convention")

# target: package
report contains violation if {
	some convention in cfg.conventions
	some target in convention.targets

	target == "package"

	not regex.match(convention.pattern, ast.package_name)

	violation := with_description(
		result.fail(rego.metadata.chain(), result.location(input["package"])),
		sprintf(
			"Naming convention violation: package name %q does not match pattern '%s'",
			[ast.package_name, convention.pattern],
		),
	)
}

# target: rule
report contains violation if {
	some convention in cfg.conventions
	some target in convention.targets

	target == "rule"

	some rule in input.rules

	not rule.head.args

	name := ast.ref_to_string(rule.head.ref)

	not regex.match(convention.pattern, name)

	violation := with_description(
		result.fail(rego.metadata.chain(), result.location(rule.head)),
		sprintf(
			"Naming convention violation: rule name %q does not match pattern '%s'",
			[name, convention.pattern],
		),
	)
}

# target: function
report contains violation if {
	some convention in cfg.conventions
	some target in convention.targets

	target == "function"

	some rule in ast.functions

	name := ast.ref_to_string(rule.head.ref)

	not regex.match(convention.pattern, name)

	violation := with_description(
		result.fail(rego.metadata.chain(), result.location(rule.head)),
		sprintf(
			"Naming convention violation: function name %q does not match pattern '%s'",
			[name, convention.pattern],
		),
	)
}

# target: var
report contains violation if {
	some convention in cfg.conventions
	some target in convention.targets

	target in {"var", "variable"}

	some var in ast.find_vars(input.rules)

	not regex.match(convention.pattern, var.value)

	violation := with_description(
		result.fail(rego.metadata.chain(), result.location(var)),
		sprintf(
			"Naming convention violation: variable name %q does not match pattern '%s'",
			[var.value, convention.pattern],
		),
	)
}

with_description(violation, description) := json.patch(
	violation,
	[{"op": "replace", "path": "/description", "value": description}],
)
