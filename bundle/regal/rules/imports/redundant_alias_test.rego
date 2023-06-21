package regal.rules.imports_test

import future.keywords.if

import data.regal.config
import data.regal.rules.imports.common_test.report

test_fail_redundant_alias if {
	r := report(`import data.foo.bar as bar`)
	r == {{
		"category": "imports",
		"description": "Redundant alias",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/redundant-alias", "imports"),
		}],
		"title": "redundant-alias",
		"location": {"col": 8, "file": "policy.rego", "row": 3, "text": "import data.foo.bar as bar"},
		"level": "error",
	}}
}

test_success_not_redundant_alias if {
	report(`import data.foo.bar as valid`) == set()
}
