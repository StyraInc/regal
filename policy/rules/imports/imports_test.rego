package regal.rules.imports_test

import future.keywords.if

import data.regal
import data.regal.rules.imports

test_fail_future_keywords_import_wildcard if {
	ast := regal.ast(`import future.keywords`)
	result := imports.violation with input as ast
	result == {{
		"description": "Use explicit future keyword imports",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-imports-001"}],
		"scope": "rule",
		"title": "STY-IMPORTS-001",
	}}
}

test_success_future_keywords_import_specific if {
	ast := regal.ast(`import future.keywords.contains`)
	result := imports.violation with input as ast
	result == set()
}

test_success_future_keywords_import_specific_many if {
	ast := regal.ast(`
    import future.keywords.contains
    import future.keywords.if
    import future.keywords.in
    `)
	result := imports.violation with input as ast
	result == set()
}
