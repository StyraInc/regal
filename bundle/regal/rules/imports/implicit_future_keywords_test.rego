package regal.rules.imports["implicit-future-keywords_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.imports["implicit-future-keywords"] as rule

test_fail_future_keywords_import_wildcard if {
	r := rule.report with input as ast.policy(`import future.keywords`)
	r == {{
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
	r := rule.report with input as ast.policy(`import future.keywords.contains`)
	r == set()
}

test_success_future_keywords_import_specific_many if {
	r := rule.report with input as ast.policy(`
    import future.keywords.contains
    import future.keywords.if
    import future.keywords.in
    `)
	r == set()
}
