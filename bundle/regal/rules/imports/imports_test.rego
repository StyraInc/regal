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
			"ref": config.docs.resolve_url("$baseUrl/$category/implicit-future-keywords", "imports"),
		}],
		"title": "implicit-future-keywords",
		"location": {"col": 8, "file": "policy.rego", "row": 3, "text": `import future.keywords`},
		"level": "error",
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
			"ref": config.docs.resolve_url("$baseUrl/$category/avoid-importing-input", "imports"),
		}],
		"title": "avoid-importing-input",
		"location": {"col": 8, "file": "policy.rego", "row": 3, "text": `import input.foo`},
		"level": "error",
	}}
}

test_fail_import_data if {
	report(`import data`) == {{
		"category": "imports",
		"description": "Redundant import of data",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/redundant-data-import", "imports"),
		}],
		"title": "redundant-data-import",
		"location": {"col": 8, "file": "policy.rego", "row": 3, "text": `import data`},
		"level": "error",
	}}
}

test_fail_import_data_aliased if {
	report(`import data as d`) == {{
		"category": "imports",
		"description": "Redundant import of data",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/redundant-data-import", "imports"),
		}],
		"title": "redundant-data-import",
		"location": {"col": 8, "file": "policy.rego", "row": 3, "text": `import data as d`},
		"level": "error",
	}}
}

test_success_import_data_path if {
	report(`import data.something`) == set()
}

test_fail_duplicate_import if {
	r := report(`
import data.foo
import data.foo
	`)
	r == {{
		"category": "imports",
		"description": "Import shadows another import",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/import-shadows-import", "imports"),
		}],
		"title": "import-shadows-import",
		"location": {"col": 8, "file": "policy.rego", "row": 5, "text": `import data.foo`},
		"level": "error",
	}}
}

test_fail_duplicate_import_alias if {
	report(`
import data.foo
import data.bar as foo
	`) == {{
		"category": "imports",
		"description": "Import shadows another import",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/import-shadows-import", "imports"),
		}],
		"title": "import-shadows-import",
		"location": {"col": 8, "file": "policy.rego", "row": 5, "text": `import data.bar as foo`},
		"level": "error",
	}}
}

report(snippet) := report if {
	# regal ignore:external-reference
	report := imports.report with input as ast.policy(snippet)
}
