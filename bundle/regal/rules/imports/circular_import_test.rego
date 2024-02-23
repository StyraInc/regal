package regal.rules.imports["circular-import_test"]

import rego.v1

import data.regal.rules.imports["circular-import"] as rule

test_aggregate_rule_empty_if_no_refs if {
	module := regal.parse_module("example.rego", `
    package policy

    allow := true
    `)

	aggregate := rule.aggregate with input as module

	aggregate == set()
}

test_aggregate_rule_empty_if_no_static_refs if {
	module := regal.parse_module("example.rego", `
    package policy

    allow := data[foo]
    `)

	aggregate := rule.aggregate with input as module

	aggregate == set()
}

test_aggregate_rule_contains_single_self_ref if {
	module := regal.parse_module("example.rego", `
    package example

    import data.example
    `)

	aggregate := rule.aggregate with input as module

	aggregate == {{
		"aggregate_data": {"refs": {{"location": {"col": 12, "row": 4}, "package_path": "data.example"}}},
		"aggregate_source": {"file": "example.rego", "package_path": ["example"]},
		"rule": {"category": "imports", "title": "circular-import"},
	}}
}

test_aggregate_rule_surfaces_refs if {
	module := regal.parse_module("example.rego", `
    package policy.foo

    import future.keywords

    import data.foo.bar

    allow := data.baz.qux

    deny contains message if {
      data.config.deny.enabled
      message := "deny"
    }
    `)

	aggregate := rule.aggregate with input as module

	aggregate == {{
		"aggregate_data": {"refs": {
			{"location": {"col": 7, "row": 11}, "package_path": "data.config.deny.enabled"},
			{"location": {"col": 12, "row": 6}, "package_path": "data.foo.bar"},
			{"location": {"col": 14, "row": 8}, "package_path": "data.baz.qux"},
		}},
		"aggregate_source": {
			"file": "example.rego",
			"package_path": ["policy", "foo"],
		},
		"rule": {"category": "imports", "title": "circular-import"},
	}}
}

test_import_graph if {
	r := rule.import_graph with input as {"aggregate": {
		{
			"aggregate_data": {"refs": {{"location": {"col": 12, "row": 3}, "package_path": "data.policy.b"}}},
			"aggregate_source": {
				"file": "a.rego",
				"package_path": ["policy", "a"],
			},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {{"location": {"col": 12, "row": 3}, "package_path": "data.policy.c"}}},
			"aggregate_source": {
				"file": "b.rego",
				"package_path": ["policy", "b"],
			},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {{"location": {"col": 12, "row": 3}, "package_path": "data.policy.a"}}},
			"aggregate_source": {
				"file": "c.rego",
				"package_path": ["policy", "c"],
			},
			"rule": {"category": "imports", "title": "circular-import"},
		},
	}}

	r == {"data.policy.a": {"data.policy.b"}, "data.policy.b": {"data.policy.c"}, "data.policy.c": {"data.policy.a"}}
}

test_import_graph_self_import if {
	r := rule.import_graph with input as {"aggregate": {{
		"aggregate_data": {"refs": {{"location": {"col": 12, "row": 4}, "package_path": "data.example"}}},
		"aggregate_source": {"file": "example.rego", "package_path": ["example"]},
		"rule": {"category": "imports", "title": "circular-import"},
	}}}

	r == {"data.example": {"data.example"}}
}

test_self_reachable if {
	r := rule.self_reachable with rule.import_graph as {
		"data.policy.a": {"data.policy.b"},
		"data.policy.b": {"data.policy.c"}, "data.policy.c": {"data.policy.a"},
	}

	r == {"data.policy.a", "data.policy.b", "data.policy.c"}
}

test_groups if {
	r := rule.groups with rule.import_graph as {
		"data.policy.a": {"data.policy.b"},
		"data.policy.b": {"data.policy.c"},
		"data.policy.c": {"data.policy.a"},
		"data.policy.d": {"data.policy.e"},
		"data.policy.e": {"data.policy.f"},
		"data.policy.f": {"data.policy.d"},
		"data.policy.g": {"data.policy.g"},
	}

	r == {
		{"data.policy.a", "data.policy.b", "data.policy.c"},
		{"data.policy.d", "data.policy.e", "data.policy.f"},
		{"data.policy.g"},
	}
}

test_groups_empty_graph if {
	r := rule.groups with rule.import_graph as {"data.policy.a": {}}

	r == set()
}

test_package_locations if {
	r := rule.package_locations with input as {"aggregate": {
		{
			"aggregate_data": {"refs": {{"location": {"col": 12, "row": 3}, "package_path": "data.policy.b"}}},
			"aggregate_source": {"file": "a.rego", "package_path": ["policy.a"]},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {{"location": {"col": 12, "row": 3}, "package_path": "data.policy.c"}}},
			"aggregate_source": {"file": "b.rego", "package_path": ["policy.b"]},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {{"location": {"col": 12, "row": 3}, "package_path": "data.policy.a"}}},
			"aggregate_source": {"file": "c.rego", "package_path": ["policy.c"]},
			"rule": {"category": "imports", "title": "circular-import"},
		},
	}}

	r == {
		"data.policy.a": {"data.policy.c": {{"col": 12, "file": "c.rego", "row": 3}}},
		"data.policy.b": {"data.policy.a": {{"col": 12, "file": "a.rego", "row": 3}}},
		"data.policy.c": {"data.policy.b": {{"col": 12, "file": "b.rego", "row": 3}}},
	}
}

test_aggregate_report_fails_when_cycle_present if {
	r := rule.aggregate_report with input as {"aggregate": {
		{
			"aggregate_data": {"refs": {{"location": {"col": 12, "row": 3}, "package_path": "data.policy.b"}}},
			"aggregate_source": {
				"file": "a.rego",
				"package_path": ["policy", "a"],
			},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {{"location": {"col": 0, "row": 2}, "package_path": "data.policy.a"}}},
			"aggregate_source": {
				"file": "b.rego",
				"package_path": ["policy", "b"],
			},
			"rule": {"category": "imports", "title": "circular-import"},
		},
	}}

	r == {{
		"category": "imports",
		"description": "Circular import detected in: data.policy.a, data.policy.b",
		"level": "error",
		"location": {"col": 0, "file": "b.rego", "row": 2},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/imports/circular-import",
		}],
		"title": "circular-import",
	}}
}

test_aggregate_report_fails_when_cycle_present_in_1_package if {
	r := rule.aggregate_report with input as {"aggregate": {{
		"aggregate_data": {"refs": {{"location": {"col": 12, "row": 3}, "package_path": "data.policy.a"}}},
		"aggregate_source": {
			"file": "a.rego",
			"package_path": ["policy", "a"],
		},
		"rule": {"category": "imports", "title": "circular-import"},
	}}}

	r == {{
		"category": "imports",
		"description": "Circular self-dependency in: data.policy.a",
		"level": "error",
		"location": {"col": 12, "file": "a.rego", "row": 3},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/imports/circular-import",
		}],
		"title": "circular-import",
	}}
}

test_aggregate_report_fails_when_cycle_present_in_n_packages if {
	r := rule.aggregate_report with input as {"aggregate": {
		{
			"aggregate_data": {"refs": {{"location": {"col": 12, "row": 3}, "package_path": "data.policy.b"}}},
			"aggregate_source": {"file": "a.rego", "package_path": ["policy", "a"]},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {{"location": {"col": 12, "row": 3}, "package_path": "data.policy.c"}}},
			"aggregate_source": {"file": "b.rego", "package_path": ["policy", "b"]},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {{"location": {"col": 12, "row": 3}, "package_path": "data.policy.a"}}},
			"aggregate_source": {"file": "c.rego", "package_path": ["policy", "c"]},
			"rule": {"category": "imports", "title": "circular-import"},
		},
	}}

	r == {{
		"category": "imports",
		"description": "Circular import detected in: data.policy.a, data.policy.b, data.policy.c",
		"level": "error",
		"location": {"col": 12, "file": "c.rego", "row": 3},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/imports/circular-import",
		}],
		"title": "circular-import",
	}}
}
