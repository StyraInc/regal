package regal.main

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config

report contains violation if {
	not is_object(input)

	violation := {
		"category": "error",
		"title": "invalid-input",
		"description": "provided input must be a JSON AST",
	}
}

report contains violation if {
	not input["package"]

	violation := {
		"category": "error",
		"title": "invalid-input",
		"description": "provided input must be a JSON AST",
	}
}

to_meta(category, title) := {
	"custom": {"category": category},
	"title": title,
}

# Check bundled rules
report contains violation if {
	some category, title
	config.merged_config.rules[category][title]

	config.for_rule(to_meta(category, title)).level != "ignore"

	violation := data.regal.rules[category][title].report[_]

	not ignored(violation, ignore_directives)
}

# Check custom rules
report contains violation if {
	some category, title

	violation := data.custom.regal.rules[category][title].report[_]

	config.for_rule(to_meta(category, title)).level != "ignore"

	not ignored(violation, ignore_directives)
}

ignored(violation, directives) if {
	ignored_rules := directives[violation.location.row]
	violation.title in ignored_rules
}

ignore_directives[row] := rules if {
	some comment in input.comments
	text := trim_space(base64.decode(comment.Text))

	startswith(text, "regal")

	i := indexof(text, "ignore:")
	i != -1

	list := regex.replace(substring(text, i + 7, -1), `\s`, "")

	row := comment.Location.row + 1
	rules := split(list, ",")
}
