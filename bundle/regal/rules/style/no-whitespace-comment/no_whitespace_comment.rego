# METADATA
# description: Comment should start with whitespace
package regal.rules.style["no-whitespace-comment"]

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	some comment in ast.comments_decoded

	not regex.match(`^[\s#]*$|^#*[\s]+.*$`, comment.text)
	not _excepted(comment.text)

	violation := result.fail(rego.metadata.chain(), result.location(comment))
}

_excepted(text) if regex.match(config.rules.style["no-whitespace-comment"]["except-pattern"], text)
