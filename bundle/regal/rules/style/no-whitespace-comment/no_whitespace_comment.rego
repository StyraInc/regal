# METADATA
# description: Comment should start with whitespace
package regal.rules.style["no-whitespace-comment"]

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	some comment in ast.comments_decoded

	not _whitespace_comment(comment.text)

	violation := result.fail(rego.metadata.chain(), result.location(comment))
}

_whitespace_comment(text) if regex.match(`^(#*)(\s+.*|$)`, text)
_whitespace_comment(text) if regex.match(config.rules.style["no-whitespace-comment"]["except-pattern"], text)
