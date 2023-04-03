package regal.rules.comments

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# For comments, OPA uses capital-cases Text and Location rather
# than text and location. As fixing this would potentially break
# things, we need to take it into consideration here.

todo_identifiers := ["todo", "TODO", "fixme", "FIXME"]

todo_pattern := sprintf(`^\s*(%s)`, [concat("|", todo_identifiers)])

# METADATA
# title: todo-comment
# description: Avoid TODO comments
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/todo-comment
# custom:
#   category: comments
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some comment in input.comments
	text := base64.decode(comment.Text)
	regex.match(todo_pattern, text)

	violation := result.fail(rego.metadata.rule(), result.location(comment))
}
