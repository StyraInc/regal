package regal.rules.comments

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal

# TODO: Normalize Text and Location -> text and location

todo_identifiers := ["todo", "TODO", "fixme", "FIXME"]
todo_pattern := sprintf("(%s)", [concat("|", todo_identifiers)])

# METADATA
# title: STY-COMMENTS-001
# description: TODO comment
# related_resources:
# - https://docs.styra.com/regal/rules/sty-comments-001
violation contains msg if {
    some comment in input.comments
    text := base64.decode(comment.Text)
    regex.match(todo_pattern, text)

    msg := regal.fail(rego.metadata.rule(), {
        "location": comment.Location
    })
}
