package regal.rules.idiomatic["no-defined-entrypoint_test"]

import rego.v1

import data.regal.config

import data.regal.rules.idiomatic["no-defined-entrypoint"] as rule

# regal ignore:rule-length
test_aggregate_entrypoints if {
	module := regal.parse_module("policy.rego", `
# METADATA
# entrypoint: true
package p

# METADATA
# entrypoint: true
allow := false`)

	aggregate := rule.aggregate with input as module
	aggregate == {
		{
			"aggregate_data": {"entrypoint": {
				"col": 1,
				"row": 2,
				"end": {
					"col": 19,
					"row": 3,
				},
				"text": "# METADATA\n# entrypoint: true",
			}},
			"aggregate_source": {"file": "policy.rego", "package_path": ["p"]},
			"rule": {"category": "idiomatic", "title": "no-defined-entrypoint"},
		},
		{
			"aggregate_data": {"entrypoint": {
				"col": 1,
				"row": 6,
				"end": {
					"col": 19,
					"row": 7,
				},
				"text": "# METADATA\n# entrypoint: true",
			}},
			"aggregate_source": {"file": "policy.rego", "package_path": ["p"]},
			"rule": {"category": "idiomatic", "title": "no-defined-entrypoint"},
		},
	}
}

test_fail_no_entrypoint_defined if {
	r := rule.aggregate_report with input as {"aggregate": set()}
	r == {{
		"category": "idiomatic",
		"description": "Missing entrypoint annotation",
		"level": "error",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/no-defined-entrypoint", "idiomatic"),
		}],
		"title": "no-defined-entrypoint",
	}}
}

test_success_single_entrypoint_defined if {
	r := rule.aggregate_report with input as {"aggregate": [{
		"aggregate_data": {"entrypoint": {"col": 1, "file": "policy.rego", "row": 2}},
		"aggregate_source": {"file": "policy.rego", "package_path": ["p"]},
		"rule": {"category": "idiomatic", "title": "no-defined-entrypoint"},
	}]}
	r == set()
}

test_success_multiple_entrypoints_defined if {
	r := rule.aggregate_report with input as {"aggregate": [
		{
			"aggregate_data": {"entrypoint": {"col": 1, "file": "policy.rego", "row": 2}},
			"aggregate_source": {"file": "policy.rego", "package_path": ["p"]},
			"rule": {"category": "idiomatic", "title": "no-defined-entrypoint"},
		},
		{
			"aggregate_data": {"entrypoint": {"col": 1, "file": "policy.rego", "row": 6}},
			"aggregate_source": {"file": "policy.rego", "package_path": ["p"]},
			"rule": {"category": "idiomatic", "title": "no-defined-entrypoint"},
		},
	]}
	r == set()
}
