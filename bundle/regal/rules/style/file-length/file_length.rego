# METADATA
# description: Max file length exceeded
package regal.rules.style["file-length"]

import data.regal.config
import data.regal.result

report contains violation if {
	cfg := config.for_rule("style", "file-length")

	count(input.regal.file.lines) > cfg["max-file-length"]

	violation := result.fail(rego.metadata.chain(), result.location(input["package"]))
}
