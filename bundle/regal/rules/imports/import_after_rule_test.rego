package regal.rules.imports["import-after-rule_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.imports["import-after-rule"] as rule

test_fail_import_after_rule if {
	module := ast.policy(`
	rule := true

	import data.foo
	`)

	r := rule.report with input as module
	r == {{
		"category": "imports",
		"description": "Import declared after rule",
		"level": "error",
		"location": {"col": 2, "file": "policy.rego", "row": 6, "text": "\timport data.foo"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/import-after-rule", "imports"),
		}],
		"title": "import-after-rule",
	}}
}

test_success_import_before_rule if {
	module := ast.policy(`
	import data.foo

	rule := true
	`)

	r := rule.report with input as module
	r == set()
}
