package regal.rules.imports["prefer-package-imports_test"]

import rego.v1

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

test_fail_aggregate_report_on_imported_rule if {
	r := rule.aggregate_report with input.aggregate as {
		{"aggregate_data": {
			"package_path": ["a"],
			"imports": [
				{
					# likely import of rule — should fail
					"path": ["b", "c"],
					"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "import data.b.c"},
				},
				# import of package, should not fail
				{"path": ["b"], "location": {"col": 1, "file": "policy.rego", "row": 4, "text": "import data.b"}},
				# unresolved import, should not fail
				{"path": ["c"], "location": {"col": 1, "file": "policy.rego", "row": 5, "text": "import data.c"}},
			],
		}},
		{"aggregate_data": {"package_path": ["b"], "imports": []}},
	}

	r == {{
		"category": "imports",
		"description": "Prefer importing packages over rules",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "import data.b.c"},
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

# unresolved imports should be flagged by an `unresolved-import`
# rule instead. see https://github.com/StyraInc/regal/issues/300
test_success_aggregate_report_on_import_with_unresolved_path if {
	r := rule.aggregate_report with input.aggregate as {
		{"aggregate_data": {
			"package_path": ["a"],
			"imports": [{
				"path": ["foo"],
				"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "import data.b"},
			}],
		}},
		{"aggregate_data": {
			"package_path": ["bar"],
			"imports": [],
		}},
	}

	r == set()
}

test_success_aggregate_report_ignored_import_path if {
	aggregate := {
		{"aggregate_data": {
			"package_path": ["a"],
			"imports": [{
				"path": ["b", "c"],
				"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "import data.b.c"},
			}],
		}},
		{"aggregate_data": {
			"package_path": ["b"],
			"imports": [],
		}},
	}

	r := rule.aggregate_report with input.aggregate as aggregate with config.for_rule as {
		"level": "error",
		"ignore-import-paths": ["data.b.c"],
	}

	r == set()
}

test_aggregate_ignores_imports_of_regal_in_custom_rule if {
	r := rule.aggregate with input as regal.parse_module("p.rego", `
	package custom.regal.rules.foo.bar

	import data.regal.ast

	import data.a.b.c
	`)

	r == {{
		"aggregate_data": {
			"imports": [{
				"location": {"col": 2, "file": "p.rego", "row": 6, "text": "\timport data.a.b.c"},
				"path": ["a", "b", "c"],
			}],
			"package_path": ["custom", "regal", "rules", "foo", "bar"],
		},
		"aggregate_source": {"file": "p.rego", "package_path": ["custom", "regal", "rules", "foo", "bar"]},
		"rule": {"category": "imports", "title": "prefer-package-imports"},
	}}
}
