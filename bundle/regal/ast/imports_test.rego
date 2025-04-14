package regal.ast_test

import data.regal.ast

test_imports_keyword_rego_v1[keyword] if {
	module := ast.policy("import rego.v1")

	some keyword in ["in", "if", "every", "contains"]

	ast.imports_keyword(module.imports, keyword)
}

test_imports_keyword_future_keywords_all[keyword] if {
	module := ast.policy("import future.keywords")

	some keyword in ["in", "if", "every", "contains"]

	ast.imports_keyword(module.imports, keyword)
}

test_imports_keyword_future_keywords_single if {
	module := ast.policy("import future.keywords.contains")

	ast.imports_keyword(module.imports, "contains")

	not ast.imports_keyword(module.imports, "in")
	not ast.imports_keyword(module.imports, "if")
	not ast.imports_keyword(module.imports, "every")
}

test_imports_keyword_future_keywords_every if {
	module := ast.policy("import future.keywords.every")

	ast.imports_keyword(module.imports, "every")
	ast.imports_keyword(module.imports, "in")

	not ast.imports_keyword(module.imports, "if")
	not ast.imports_keyword(module.imports, "contains")
}

test_imports_all_kewords_opa_1dot0[keyword] if {
	some keyword in ["in", "if", "every", "contains"]

	ast.imports_keyword([], keyword) with input.regal.file.rego_version as "v1"
		with data.regal.capabilities.is_opa_v1 as true
}
