package regal.rules.bugs["annotation-without-metadata_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["annotation-without-metadata"] as rule

test_fail_annotation_without_metadata if {
	module := ast.with_rego_v1(`
# title: allow
allow := false
	`)
	r := rule.report with input as module

	r == {{
		"category": "bugs",
		"description": "Annotation without metadata",
		"level": "error",
		"location": {
			"col": 1,
			"file": "policy.rego",
			"row": 6,
			"end": {
				"col": 15,
				"row": 6,
			},
			"text": "# title: allow",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/annotation-without-metadata", "bugs"),
		}],
		"title": "annotation-without-metadata",
	}}
}

test_success_annotation_with_metadata if {
	module := ast.with_rego_v1(`
# METADATA
# title: allow
allow := false
	`)
	r := rule.report with input as module

	r == set()
}

test_success_annotation_but_no_metadata_location if {
	module := ast.with_rego_v1(`
allow := false # title: allow
	`)
	r := rule.report with input as module

	r == set()
}

test_success_annotation_without_metadata_but_comment_preceding if {
	module := ast.with_rego_v1(`
# something that is not an annotation here will cancel this rule
# as this is less likely to be a mistake... but weird
# title: allow
allow := false
	`)
	r := rule.report with input as module

	r == set()
}
