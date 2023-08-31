# METADATA
# description: Max file length exceeded
package regal.rules.style["file-length"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

cfg := config.for_rule("style", "file-length")

report contains violation if {
	count(input.regal.file.lines) > cfg["max-file-length"]

	violation := result.fail(rego.metadata.chain(), result.location(input["package"]))
}
