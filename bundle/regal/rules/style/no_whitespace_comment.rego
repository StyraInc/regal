# METADATA
# description: Comment should start with whitespace
package regal.rules.style["no-whitespace-comment"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	some comment in ast.comments_decoded

	not _whitespace_comment(comment.Text)

	violation := result.fail(rego.metadata.chain(), result.location(comment))
}

_whitespace_comment(text) if startswith(text, " ")

_whitespace_comment(text) if text in {"", "\n"}

_whitespace_comment(text) if startswith(text, "##")
