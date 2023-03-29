package regal.rules.comments

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# TODO: Normalize Text and Location -> text and location

todo_identifiers := ["todo", "TODO", "fixme", "FIXME"]

todo_pattern := sprintf("(%s)", [concat("|", todo_identifiers)])

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
