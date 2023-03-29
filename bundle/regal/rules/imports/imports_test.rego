package regal.rules.imports_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.imports

test_fail_future_keywords_import_wildcard if {
	report(`import future.keywords`) == {{
		"category": "imports",
		"description": "Use explicit future keyword imports",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/implicit-future-keywords",
		}],
		"title": "implicit-future-keywords",
		"location": {"col": 8, "file": "policy.rego", "row": 8},
	}}
}

test_success_future_keywords_import_specific if {
	report(`import future.keywords.contains`) == set()
}

test_success_future_keywords_import_specific_many if {
	report(`
    import future.keywords.contains
    import future.keywords.if
    import future.keywords.in
    `) == set()
}

test_fail_import_input if {
	report(`import input.foo`) == {{
		"category": "imports",
		"description": "Avoid importing input",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/avoid-importing-input",
		}],
		"title": "avoid-importing-input",
		"location": {"col": 8, "file": "policy.rego", "row": 8},
	}}
}

test_fail_import_data if {
	report(`import data`) == {{
		"category": "imports",
		"description": "Redundant import of data",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/redundant-data-import",
		}],
		"title": "redundant-data-import",
		"location": {"col": 8, "file": "policy.rego", "row": 8},
	}}
}

test_fail_import_data_aliased if {
	report(`import data as d`) == {{
		"category": "imports",
		"description": "Redundant import of data",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/redundant-data-import",
		}],
		"title": "redundant-data-import",
		"location": {"col": 8, "file": "policy.rego", "row": 8},
	}}
}

test_success_import_data_path if {
	report(`import data.something`) == set()
}

report(snippet) := report if {
	report := imports.report with input as ast.with_future_keywords(snippet)
		with config.for_rule as {"enabled": true}
}
