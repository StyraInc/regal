# METADATA
# description: Importing own package is pointless
package regal.rules.imports["pointless-import"]

import data.regal.ast
import data.regal.result

report contains violation if {
	plen := count(input.package.path)
	path := input.imports[_].path
	ilen := count(path.value)

	# allow package a.b to import a.b.c.d.e but not a.b or a.b.c
	ilen - plen < 2

	same := array.slice(path.value, 0, plen)

	ast.ref_value_equal(input.package.path, same)

	violation := result.fail(rego.metadata.chain(), result.location(path))
}
