# METADATA
# description: All I ever do is print
# related_resources:
# - description: documentation
#   ref: https://www.acmecorp.example.org/docs/regal/package
package custom.regal.rules.utils["printer"]

import rego.v1

report contains "never happens" if {
	print(input.regal.file.name)

	false
}
