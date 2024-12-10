# METADATA
# description: various helpers to be used for testing completions providers
package regal.lsp.completion.providers.test_utils

# METADATA
# description: returns a map of all parsed modules in the workspace
parsed_modules(workspace) := {file_uri: parsed_module |
	some file_uri, contents in workspace
	parsed_module := regal.parse_module(file_uri, contents)
}

# METADATA
# description: adds location metadata to provided module, to be used as input
input_module_with_location(module, policy, location) := object.union(module, {"regal": {
	"file": {
		"name": "p.rego",
		"lines": split(policy, "\n"),
	},
	"context": {"location": location},
}})

# METADATA
# description: same as input_module_with_location, but accepts text content rather than a module
input_with_location(policy, location) := {"regal": {
	"file": {
		"name": "p.rego",
		"lines": split(policy, "\n"),
	},
	"context": {"location": location},
}}
