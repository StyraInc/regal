package regal.rules.imports_test

import future.keywords.if

import data.regal
import data.regal.rules.imports

test_fail_future_keywords_import_wildcard if {
	report(`import future.keywords`) == {{
		"category": "imports",
		"description": "Use explicit future keyword imports",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/implicit-future-keywords"
		}],
		"title": "implicit-future-keywords",
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
			"ref": "https://docs.styra.com/regal/rules/avoid-importing-input"
		}],
		"title": "avoid-importing-input",
	}}
}

report(snippet) := report {
	report := imports.report with input as regal.ast(snippet) with regal.rule_config as {"enabled": true}
}
