package regal.rules.imports["circular-import_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.imports["circular-import"] as rule

test_aggregate_rule_empty_if_no_refs if {
	aggregate := rule.aggregate with input as ast.policy("allow := true")

	aggregate == set()
}

test_aggregate_rule_empty_if_no_static_refs if {
	aggregate := rule.aggregate with input as ast.policy("allow := data[foo]")

	aggregate == set()
}

test_aggregate_rule_contains_single_self_ref if {
	aggregate := rule.aggregate with input as ast.policy("import data.example")

	aggregate == {{
		"aggregate_data": {"refs": {["data.example", "3:8:3:20"]}},
		"aggregate_source": {"file": "policy.rego", "package_path": ["policy"]},
		"rule": {"category": "imports", "title": "circular-import"},
	}}
}

test_aggregate_rule_surfaces_refs if {
	aggregate := rule.aggregate with input as regal.parse_module("example.rego", `
    package policy.foo

    import future.keywords

    import data.foo.bar

    allow := data.baz.qux

    deny contains message if {
      data.config.deny.enabled
      message := "deny"
    }
    `)

	aggregate == {{
		"aggregate_data": {"refs": {
			["data.config.deny.enabled", "11:7:11:31"],
			["data.foo.bar", "6:12:6:24"],
			["data.baz.qux", "8:14:8:26"],
		}},
		"aggregate_source": {
			"file": "example.rego",
			"package_path": ["policy", "foo"],
		},
		"rule": {"category": "imports", "title": "circular-import"},
	}}
}

test_import_graph if {
	r := rule._import_graph with input as {"aggregate": {
		{
			"aggregate_data": {"refs": {["data.policy.b", "3:12:3:12"]}},
			"aggregate_source": {
				"file": "a.rego",
				"package_path": ["policy", "a"],
			},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {["data.policy.c", "3:12:3:12"]}},
			"aggregate_source": {
				"file": "b.rego",
				"package_path": ["policy", "b"],
			},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {["data.policy.a", "3:12:3:12"]}},
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
	r := rule._import_graph with input as {"aggregate": {{
		"aggregate_data": {"refs": {["data.example", "4:12:4:12"]}},
		"aggregate_source": {"file": "example.rego", "package_path": ["example"]},
		"rule": {"category": "imports", "title": "circular-import"},
	}}}

	r == {"data.example": {"data.example"}}
}

test_self_reachable if {
	r := rule._self_reachable with rule._import_graph as {
		"data.policy.a": {"data.policy.b"},
		"data.policy.b": {"data.policy.c"}, "data.policy.c": {"data.policy.a"},
	}

	r == {"data.policy.a", "data.policy.b", "data.policy.c"}
}

test_groups if {
	r := rule._groups with rule._import_graph as {
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
	r := rule._groups with rule._import_graph as {"data.policy.a": {}}

	r == set()
}

test_package_locations if {
	r := rule._package_locations with input as {"aggregate": {
		{
			"aggregate_data": {"refs": {["data.policy.b", "3:12:3:12"]}},
			"aggregate_source": {"file": "a.rego", "package_path": ["policy.a"]},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {["data.policy.c", "3:12:3:12"]}},
			"aggregate_source": {"file": "b.rego", "package_path": ["policy.b"]},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {["data.policy.a", "3:12:3:12"]}},
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
			"aggregate_data": {"refs": {["data.policy.b", "3:12:3:12"]}},
			"aggregate_source": {
				"file": "a.rego",
				"package_path": ["policy", "a"],
			},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {["data.policy.a", "2:0:2:0"]}},
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
			"ref": config.docs.resolve_url("$baseUrl/$category/circular-import", "imports"),
		}],
		"title": "circular-import",
	}}
}

test_aggregate_report_fails_when_cycle_present_in_1_package if {
	r := rule.aggregate_report with input as {"aggregate": {{
		"aggregate_data": {"refs": {["data.policy.a", "3:12:3:12"]}},
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
			"ref": config.docs.resolve_url("$baseUrl/$category/circular-import", "imports"),
		}],
		"title": "circular-import",
	}}
}

test_aggregate_report_fails_when_cycle_present_in_n_packages if {
	r := rule.aggregate_report with input as {"aggregate": {
		{
			"aggregate_data": {"refs": {["data.policy.b", "3:12:3:12"]}},
			"aggregate_source": {"file": "a.rego", "package_path": ["policy", "a"]},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {["data.policy.c", "3:12:3:12"]}},
			"aggregate_source": {"file": "b.rego", "package_path": ["policy", "b"]},
			"rule": {"category": "imports", "title": "circular-import"},
		},
		{
			"aggregate_data": {"refs": {["data.policy.a", "3:12:3:12"]}},
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
			"ref": config.docs.resolve_url("$baseUrl/$category/circular-import", "imports"),
		}],
		"title": "circular-import",
	}}
}
