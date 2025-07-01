# METADATA
# description: Custom rule description
package custom.regal.rules.naming.myrule

import rego.v1

import data.regal.result

report contains result.fail(rego.metadata.chain(), result.location(input.package.path[1]))
