package regal.rules.imports["prefer-package-imports_test"]

import future.keywords.if
import future.keywords.in

import data.regal.config

import data.regal.rules.imports["prefer-package-imports"] as rule

test_aggregate_collects_imports_with_location if {
	r := rule.aggregate with input as regal.parse_module("p.rego", `
	package a

	import data.b
	import data.c.d
	`)

	r == {{
		"aggregate_data": {
			"imports": [
				{"location": {"col": 2, "file": "p.rego", "row": 4, "text": "\timport data.b"}, "path": ["b"]},
				{"location": {"col": 2, "file": "p.rego", "row": 5, "text": "\timport data.c.d"}, "path": ["c", "d"]},
			],
			"package_path": ["a"],
		},
		"aggregate_source": {"file": "p.rego", "package_path": ["a"]},
		"rule": {"category": "imports", "title": "prefer-package-imports"},
	}}
}

test_fail_aggregate_report_on_import_without_matching_package if {
	r := rule.aggregate_report with input.aggregate as {{"aggregate_data": {
		"package_path": ["a"],
		"imports": [{"path": ["b"], "location": {"col": 1, "file": "policy.rego", "row": 3, "text": "import data.b"}}],
	}}}

	r == {{
		"category": "imports",
		"description": "Prefer importing packages over rules",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "import data.b"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/prefer-package-imports", "imports"),
		}],
		"title": "prefer-package-imports",
	}}
}

test_success_aggregate_report_on_import_with_matching_package if {
	r := rule.aggregate_report with input.aggregate as {
		{"aggregate_data": {
			"package_path": ["a"],
			"imports": [{
				"path": ["b"],
				"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "import data.b"},
			}],
		}},
		{"aggregate_data": {
			"package_path": ["b"],
			"imports": [],
		}},
	}

	r == set()
}

test_success_aggregate_report_ignored_import_path if {
	aggregate := {{"aggregate_data": {
		"package_path": ["a"],
		"imports": [{"path": ["b"], "location": {"col": 1, "file": "policy.rego", "row": 3, "text": "import data.b"}}],
	}}}

	r := rule.aggregate_report with input.aggregate as aggregate with config.for_rule as {
		"level": "error",
		"ignore-import-paths": ["data.b"],
	}

	r == set()
}
