# METADATA
# description: Avoid importing input
package regal.rules.imports["avoid-importing-input"]

import data.regal.result

report contains violation if {
	some imported in input.imports

	imported.path.value[0].value == "input"

	# Allow aliasing input, eg `import input as tfplan`:
	not _aliased_input(imported)

	violation := result.fail(rego.metadata.chain(), result.location(imported.path.value[0]))
}

_aliased_input(imported) if {
	count(imported.path.value) == 1
	imported.alias
}
