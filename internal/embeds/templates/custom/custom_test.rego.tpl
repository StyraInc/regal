package custom.regal.rules.{{.Category}}{{.NameTest}}

import future.keywords.if
import future.keywords.in

import data.custom.regal.rules.{{.Category}}{{.Name}} as rule

# Example test, replace with your own
test_rule_named_foo_not_allowed {
    module := regal.parse_module("example.rego", `
    package policy

    foo := true`)

    r := rule.report with input as module

    # Use print(r) here to see the report. Great for development!

    r == {{ "{{" }}
    	"category": "{{.Category}}",
    	"description": "Add description of rule here!",
    	"level": "error",
    	"location": {"col": 5, "file": "example.rego", "row": 4, "text": "    foo := true"},
    	"title": "{{.NameOriginal}}"
    }}
}
