package regal.rules.style["detached-metadata_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style["detached-metadata"] as rule

test_fail_detached_metadata if {
	r := rule.report with input as regal.parse_module("p.rego", `
package p

# METADATA
# description: allow is always true

allow := true
`)
	r == {{
		"category": "style",
		"description": "Detached metadata annotation",
		"level": "error",
		"location": {"col": 1, "file": "p.rego", "row": 4, "text": "# METADATA"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/detached-metadata", "style"),
		}],
		"title": "detached-metadata",
	}}
}

test_fail_very_detached_metadata if {
	r := rule.report with input as regal.parse_module("p.rego", `
package p

# METADATA
# description: allow is always true



allow := true
`)
	r == {{
		"category": "style",
		"description": "Detached metadata annotation",
		"level": "error",
		"location": {"col": 1, "file": "p.rego", "row": 4, "text": "# METADATA"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/detached-metadata", "style"),
		}],
		"title": "detached-metadata",
	}}
}

test_success_attached_metadata if {
	r := rule.report with input as ast.policy(`
# METADATA
# title: valid
allow := true
`)
	r == set()
}

test_success_detached_document_scope_ok if {
	r := rule.report with input as regal.parse_module("p.rego", `
package p

# METADATA
# scope: document
# descriptiom: allow allows

# METADATA
# title: allow
allow := true
`)
	r == set()
}
