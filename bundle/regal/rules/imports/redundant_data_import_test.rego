package regal.rules.imports_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.imports
import data.regal.rules.imports.common_test.report

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
