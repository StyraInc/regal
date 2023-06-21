package regal.rules.imports_test

import future.keywords.if

import data.regal.config
import data.regal.rules.imports.common_test.report

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

test_sucess_import_aliased_input if {
	report(`import input as tfplan`) == set()
}

test_fail_import_input_aliased_attribute if {
	report(`import input.foo.bar as barbar`) == {{
		"category": "imports",
		"description": "Avoid importing input",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/avoid-importing-input", "imports"),
		}],
		"title": "avoid-importing-input",
		"location": {"col": 8, "file": "policy.rego", "row": 3, "text": `import input.foo.bar as barbar`},
		"level": "error",
	}}
}
