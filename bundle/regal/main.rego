package regal.main

import future.keywords.contains
import future.keywords.if
import future.keywords.in

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

report contains violation if {
	violation := data.regal.rules[_].report[_]
}

report contains violation if {
	violation := data.custom.regal.rules[_].report[_]
}
