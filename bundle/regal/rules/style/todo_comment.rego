# METADATA
# description: Avoid TODO comments
package regal.rules.style["todo-comment"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

# For comments, OPA uses capital-cases Text and Location rather
# than text and location. As fixing this would potentially break
# things, we need to take it into consideration here.

todo_identifiers := ["todo", "TODO", "fixme", "FIXME"]

todo_pattern := sprintf(`^\s*(%s)`, [concat("|", todo_identifiers)])

report contains violation if {
	some comment in input.comments
	text := base64.decode(comment.Text)
	regex.match(todo_pattern, text)

	violation := result.fail(rego.metadata.chain(), result.location(comment))
}
