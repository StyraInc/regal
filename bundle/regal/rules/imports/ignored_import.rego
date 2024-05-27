# METADATA
# description: Reference ignores import
package regal.rules.imports["ignored-import"]

import rego.v1

import data.regal.ast
import data.regal.result

import_paths contains path if {
	some imp in input.imports
	path := [p.value | some p in imp.path.value]

	path[0] in {"data", "input"}
	count(path) > 1
}

report contains violation if {
	some ref in ast.all_rules_refs

	ref.value[0].type == "var"
	ref.value[0].value in {"data", "input"}

	most_specific_match := regal.last(sort([ip |
		ref_path := [p.value | some p in ref.value]

		some ip in import_paths
		array.slice(ref_path, 0, count(ip)) == ip
	]))

	violation := result.fail(rego.metadata.chain(), object.union(
		result.location(ref),
		{"description": sprintf("Reference ignores import of %s", [concat(".", most_specific_match)])},
	))
}
