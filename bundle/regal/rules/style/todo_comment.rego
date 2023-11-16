# METADATA
# description: Avoid TODO comments
package regal.rules.style["todo-comment"]

import rego.v1

import data.regal.ast
import data.regal.result

# For comments, OPA uses capital-cases Text and Location rather
# than text and location. As fixing this would potentially break
# things, we need to take it into consideration here.

todo_identifiers := ["todo", "TODO", "fixme", "FIXME"]

todo_pattern := sprintf(`^\s*(%s)`, [concat("|", todo_identifiers)])

report contains violation if {
	some comment in ast.comments_decoded
	regex.match(todo_pattern, comment.Text)

	violation := result.fail(rego.metadata.chain(), result.location(comment))
}
