# METADATA
# description: Naming convention violation
package regal.rules.custom["naming-convention"]

import data.regal.ast
import data.regal.config
import data.regal.result

# target: package
report contains violation if {
	some convention in config.rules.custom["naming-convention"].conventions
	"package" in convention.targets

	not regex.match(convention.pattern, ast.package_name)

	violation := _with_description(
		result.fail(rego.metadata.chain(), result.location(input.package)),
		sprintf(
			"Naming convention violation: package name %q does not match pattern '%s'",
			[ast.package_name, convention.pattern],
		),
	)
}

# target: rule
report contains violation if {
	some convention in config.rules.custom["naming-convention"].conventions
	"rule" in convention.targets

	some rule in ast.rules

	name := ast.ref_to_string(rule.head.ref)

	not regex.match(convention.pattern, name)

	violation := _with_description(
		result.fail(rego.metadata.chain(), result.location(rule.head)),
		sprintf(
			"Naming convention violation: rule name %q does not match pattern '%s'",
			[name, convention.pattern],
		),
	)
}

# target: function
report contains violation if {
	some convention in config.rules.custom["naming-convention"].conventions
	"function" in convention.targets

	some rule in ast.functions

	name := ast.ref_to_string(rule.head.ref)

	not regex.match(convention.pattern, name)

	violation := _with_description(
		result.fail(rego.metadata.chain(), result.location(rule.head)),
		sprintf(
			"Naming convention violation: function name %q does not match pattern '%s'",
			[name, convention.pattern],
		),
	)
}

# target: var
report contains violation if {
	some convention in config.rules.custom["naming-convention"].conventions
	some target in convention.targets

	target in {"var", "variable"}

	var := ast.found.vars[_][_][_]

	not regex.match(convention.pattern, var.value)

	violation := _with_description(
		result.fail(rego.metadata.chain(), result.location(var)),
		sprintf(
			"Naming convention violation: variable name %q does not match pattern '%s'",
			[var.value, convention.pattern],
		),
	)
}

_with_description(violation, description) := json.patch(
	violation,
	[{"op": "replace", "path": "/description", "value": description}],
)
