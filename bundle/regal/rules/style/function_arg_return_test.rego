package regal.rules.style["function-arg-return_test"]

import future.keywords.if

import data.regal.ast
import data.regal.capabilities
import data.regal.config
import data.regal.rules.style["function-arg-return"] as rule

test_fail_function_arg_return_value if {
	r := rule.report with input as ast.policy(`foo := i { indexof("foo", "o", i) }`)
		with config.for_rule as {"level": "error"}
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == {{
		"category": "style",
		"description": "Function argument used for return value",
		"level": "error",
		"location": {"col": 32, "file": "policy.rego", "row": 3, "text": "foo := i { indexof(\"foo\", \"o\", i) }"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/function-arg-return", "style"),
		}],
		"title": "function-arg-return",
	}}
}

test_fail_function_arg_return_value_multi_part_ref if {
	r := rule.report with input as ast.policy(`foo := r { regex.match("foo", "foo", r) }`)
		with config.for_rule as {"level": "error"}
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == {{
		"category": "style",
		"description": "Function argument used for return value",
		"level": "error",
		"location": {"col": 38, "file": "policy.rego", "row": 3, "text": `foo := r { regex.match("foo", "foo", r) }`},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/function-arg-return", "style"),
		}],
		"title": "function-arg-return",
	}}
}

test_success_function_arg_return_value_except_function if {
	r := rule.report with input as ast.with_future_keywords(`foo := i { indexof("foo", "o", i) }`)
		with config.for_rule as {
			"level": "error",
			"except-functions": ["indexof"],
		}
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}
