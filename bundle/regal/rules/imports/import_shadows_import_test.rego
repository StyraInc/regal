package regal.rules.imports_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.imports
import data.regal.rules.imports.common_test.report

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
