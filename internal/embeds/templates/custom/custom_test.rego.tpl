package custom.regal.rules.{{.Category}}{{.NameTest}}

import rego.v1

import data.custom.regal.rules.{{.Category}}{{.Name}} as rule

# Example test, replace with your own
test_rule_named_foo_not_allowed if {
	module := regal.parse_module("example.rego", `
	package policy

	foo := true`)

	r := rule.report with input as module

	# Use print(r) here to see the report. Great for development!

	r == {{ "{{" }}
		"category": "{{.Category}}",
		"description": "Add description of rule here!",
		"level": "error",
		"location": {
			"file": "example.rego",
			"row": 4,
			"col": 2,
			"end": {
				"row": 4,
				"col": 13,
			},
			"text": "\tfoo := true"
		},
		"title": "{{.NameOriginal}}",
	}}
}
