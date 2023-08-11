package regal.rules.idiomatic["use-some-for-output-vars_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.idiomatic["use-some-for-output-vars"] as rule

test_fail_output_var_not_declared if {
	r := rule.report with input as ast.policy(`allow {
		"admin" == input.user.roles[i]
	}`)
	r == {{
		"category": "idiomatic",
		"description": "Use `some` to declare output variables",
		"level": "error",
		"location": {"col": 31, "file": "policy.rego", "row": 4, "text": "\t\t\"admin\" == input.user.roles[i]"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-some-for-output-vars", "idiomatic"),
		}],
		"title": "use-some-for-output-vars",
	}}
}

test_fail_multiple_output_vars_not_declared if {
	r := rule.report with input as ast.policy(`allow {
		foo := input.foo[i].bar[j]
	}`)
	r == {
		{
			"category": "idiomatic",
			"description": "Use `some` to declare output variables",
			"level": "error",
			"location": {"col": 20, "file": "policy.rego", "row": 4, "text": "\t\tfoo := input.foo[i].bar[j]"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/use-some-for-output-vars", "idiomatic"),
			}],
			"title": "use-some-for-output-vars",
		},
		{
			"category": "idiomatic",
			"description": "Use `some` to declare output variables",
			"level": "error",
			"location": {"col": 27, "file": "policy.rego", "row": 4, "text": "\t\tfoo := input.foo[i].bar[j]"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/use-some-for-output-vars", "idiomatic"),
			}],
			"title": "use-some-for-output-vars",
		},
	}
}

test_fail_only_one_declared if {
	r := rule.report with input as ast.policy(`allow {
		some i
		foo := input.foo[i].bar[j]
	}`)

	r == {{
		"category": "idiomatic",
		"description": "Use `some` to declare output variables",
		"level": "error",
		"location": {"col": 27, "file": "policy.rego", "row": 5, "text": "\t\tfoo := input.foo[i].bar[j]"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-some-for-output-vars", "idiomatic"),
		}],
		"title": "use-some-for-output-vars",
	}}
}

test_success_uses_some if {
	r := rule.report with input as ast.policy(`allow {
		some i
		"admin" == input.user.roles[i]
	}`)
	r == set()
}

test_success_some_iteration if {
	rule.report with input as ast.with_future_keywords(`allow {
		some i in input
		foo[i]
	}`) == set()

	rule.report with input as ast.with_future_keywords(`allow {
		some i, x in input
		input.user.roles[i]
	}`) == set()

	rule.report with input as ast.with_future_keywords(`allow {
		some x, i in input
		input.user.roles[i]
	}`) == set()

	rule.report with input as ast.with_future_keywords(`allow {
		some x, i in input
		input.user.roles[x][i]
	}`) == set()

	rule.report with input as ast.with_future_keywords(`allow {
		some i in input
		input.user.roles[_]
	}`) == set()
}

test_success_not_an_output_var if {
	r := rule.report with input as ast.policy(`
		i := 0

		allow {
			# i now an *input* var
			"admin" == input.user.roles[i]
		}`)
	r == set()
}
