package regal.rules.bugs["unused-output-variable_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["unused-output-variable"] as rule

test_fail_unused_output_variable if {
	module := ast.with_rego_v1(`
	fail if {
		input[x]
	}
	`)

	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Unused output variable",
		"level": "error",
		"location": {"col": 9, "end": {"col": 10, "row": 7}, "file": "policy.rego", "row": 7, "text": "\t\tinput[x]"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/unused-output-variable", "bugs"),
		}],
		"title": "unused-output-variable",
	}}
}

test_fail_unused_output_variable_some if {
	module := ast.with_rego_v1(`
	fail if {
		some xx
		input[xx]
	}
	`)

	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Unused output variable",
		"level": "error",
		"location": {"col": 9, "end": {"col": 11, "row": 8}, "file": "policy.rego", "row": 8, "text": "\t\tinput[xx]"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/unused-output-variable", "bugs"),
		}],
		"title": "unused-output-variable",
	}}
}

test_success_unused_wildcard if {
	module := ast.with_rego_v1(`
	success if {
		input[_]
	}
	`)

	r := rule.report with input as module
	r == set()
}

test_success_not_unused_variable_in_head_value if {
	module := ast.with_rego_v1(`
	success := x if {
		input[x]
	}
	`)

	r := rule.report with input as module
	r == set()
}

test_success_not_unused_variable_in_head_term_value if {
	module := ast.with_rego_v1(`
	success := {x} if {
		input[x]
	}
	`)

	r := rule.report with input as module
	r == set()
}

test_success_not_unused_variable_in_head_term_key if {
	module := ast.with_rego_v1(`
	success contains {x} if {
		input[x]
	}
	`)

	r := rule.report with input as module
	r == set()
}

test_success_not_unused_variable_in_head_key if {
	module := ast.with_rego_v1(`
	success contains x if {
		input[x]
	}
	`)

	r := rule.report with input as module
	r == set()
}

test_success_not_unused_output_variable_function_call if {
	module := ast.with_rego_v1(`
	success if {
		some x
		input[x]
		regex.match("[x]", x)
	}
	`)

	r := rule.report with input as module
	r == set()
}

test_success_not_unused_output_variable_function_call_arg_term if {
	module := ast.with_rego_v1(`
	success if {
		some x
		input[x]
		f({x})
	}
	`)

	r := rule.report with input as module
	r == set()
}

test_success_not_unused_output_variable_other_ref if {
	module := ast.with_rego_v1(`
	success if {
		some x
		input[x] == data.foo[x]
	}
	`)

	r := rule.report with input as module
	r == set()
}

test_success_not_unused_output_variable_head_ref if {
	module := ast.with_rego_v1(`
	success[x] if {
		some x
		input[x]
	}
	`)

	r := rule.report with input as module
	r == set()
}

test_success_not_output_variable_rule if {
	module := ast.with_rego_v1(`
	x := 1

	success := x if {
		input[x]
	}
	`)

	r := rule.report with input as module
	r == set()
}

test_success_not_output_variable_argument if {
	module := ast.with_rego_v1(`
	success(x) if {
		input[x]
	}
	`)

	r := rule.report with input as module
	r == set()
}
