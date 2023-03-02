package regal.rules.comments

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal

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
	regal.rule_config(rego.metadata.rule()).enabled == true

    some comment in input.comments
    text := base64.decode(comment.Text)
    regex.match(todo_pattern, text)

    violation := regal.fail(rego.metadata.rule(), {"location": comment.Location})
}
