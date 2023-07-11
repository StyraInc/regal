# METADATA
# description: Custom rule description
package custom.regal.rules.naming.myrule

import data.regal.result
import future.keywords.contains

report contains result.fail(rego.metadata.chain(), result.location(input["package"].path[1]))
