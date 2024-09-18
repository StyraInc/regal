# METADATA
# description: Avoid TODO comments
package regal.rules.style["todo-comment"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	todo_identifiers := ["todo", "TODO", "fixme", "FIXME"]
	todo_pattern := sprintf(`^\s*(%s)`, [concat("|", todo_identifiers)])

	some comment in ast.comments_decoded
	regex.match(todo_pattern, comment.text)

	violation := result.fail(rego.metadata.chain(), result.location(comment))
}
