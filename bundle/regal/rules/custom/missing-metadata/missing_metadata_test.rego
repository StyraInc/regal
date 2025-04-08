package regal.rules.custom["missing-metadata_test"]

import data.regal.config

import data.regal.rules.custom["missing-metadata"] as rule

test_success_aggregate_format_as_expected if {
	module := regal.parse_module("p.rego", `# METADATA
# title: pkg
package foo.bar

# METADATA
# title: rule
rule := true

none := false
`)
	aggregated := rule.aggregate with input as module

	aggregated == {{
		"aggregate_data": {
			"package_annotated": true,
			"package_location": {
				"col": 1,
				"row": 3,
				"end": {
					"col": 8,
					"row": 3,
				},
				"text": "package",
			},
			"rule_annotations": {
				"foo.bar.none": {false},
				"foo.bar.rule": {true},
			},
			"rule_locations": {"foo.bar.none": {
				"col": 1,
				"row": 9,
				"end": {
					"col": 5,
					"row": 9,
				},
				"text": "none := false",
			}},
		},
		"aggregate_source": {
			"file": "p.rego",
			"package_path": ["foo", "bar"],
		},
		"rule": {
			"category": "custom",
			"title": "missing-metadata",
		},
	}}
}

test_success_not_missing_package_metadata_report if {
	module := regal.parse_module("p.rego", `# METADATA
# title: pkg
package foo.bar
`)
	aggregated := rule.aggregate with input as module
	r := rule.aggregate_report with input.aggregate as aggregated

	r == set()
}

test_fail_missing_package_metadata_report if {
	module := regal.parse_module("p.rego", "package foo.bar")
	aggregated := rule.aggregate with input as module
	r := rule.aggregate_report with input.aggregate as aggregated

	r == {{
		"category": "custom",
		"description": "Package or rule missing metadata",
		"level": "error",
		"location": {
			"col": 1,
			"row": 1,
			"end": {
				"col": 8,
				"row": 1,
			},
			"text": "package",
			"file": "p.rego",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/missing-metadata", "custom"),
		}],
		"title": "missing-metadata",
	}}
}

test_success_one_missing_one_found_package_metadata_report if {
	module1 := regal.parse_module("p.rego", "package foo.bar")
	module2 := regal.parse_module("p.rego", `# METADATA
# title: pkg
package foo.bar
`)
	agg1 := rule.aggregate with input as module1
	agg2 := rule.aggregate with input as module2

	r := rule.aggregate_report with input.aggregate as {agg1, agg2}

	r == set()
}

test_success_not_missing_rule_metadata_report if {
	module := regal.parse_module("p.rego", `# METADATA
# title: pkg
package foo.bar

# METADATA
# title: baz
baz := true
`)

	a := rule.aggregate with input as module
	r := rule.aggregate_report with input.aggregate as a

	r == set()
}

test_fail_missing_rule_metadata_report if {
	module := regal.parse_module("p.rego", `# METADATA
# title: pkg
package foo.bar

baz := true
`)
	a := rule.aggregate with input as module
	r := rule.aggregate_report with input.aggregate as a with config.for_rule as {"level": "error"}

	r == {{
		"category": "custom",
		"description": "Package or rule missing metadata",
		"level": "error",
		"location": {
			"col": 1,
			"file": "p.rego",
			"row": 5,
			"end": {
				"col": 4,
				"row": 5,
			},
			"text": "baz := true",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/missing-metadata", "custom"),
		}],
		"title": "missing-metadata",
	}}
}

test_success_one_missing_one_found_rule_metadata_report if {
	module1 := regal.parse_module("p.rego", `# METADATA
# title: pkg
package foo.bar

rule := true
`)
	module2 := regal.parse_module("p.rego", `
package foo.bar

# METADATA
# description: hey
rule := false
`)
	agg1 := rule.aggregate with input as module1
	agg2 := rule.aggregate with input as module2

	r := rule.aggregate_report with input.aggregate as {agg1, agg2}

	r == set()
}
