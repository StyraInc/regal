package regal.rules.style["prefer-some-in-iteration_test"]

import data.regal.ast
import data.regal.capabilities
import data.regal.config
import data.regal.rules.style["prefer-some-in-iteration"] as rule

test_fail_simple_iteration if {
	policy := ast.with_rego_v1(`allow if {
		var := input.foo[_]
	}`)

	r := rule.report with config.rules as allow_nesting(2)
		with input as policy
		with config.capabilities as capabilities.provided

	r == with_location({
		"col": 10,
		"file": "policy.rego",
		"row": 6,
		"end": {
			"col": 22,
			"row": 6,
		},
		"text": "\t\tvar := input.foo[_]",
	})
}

test_fail_simple_iteration_comprehension if {
	policy := ast.with_rego_v1(`s := {p |
		p := input.foo[_]
	}`)

	r := rule.report with config.rules as allow_nesting(2)
		with input as policy
		with config.capabilities as capabilities.provided

	r == with_location({
		"col": 8,
		"file": "policy.rego",
		"row": 6,
		"end": {
			"col": 20,
			"row": 6,
		},
		"text": "\t\tp := input.foo[_]",
	})
}

test_fail_simple_iteration_output_var if {
	policy := ast.with_rego_v1(`allow if {
		input.foo[x]
	}`)

	r := rule.report with config.rules as allow_nesting(2)
		with input as policy
		with config.capabilities as capabilities.provided

	r == with_location({
		"col": 3,
		"file": "policy.rego",
		"row": 6,
		"end": {
			"col": 15,
			"row": 6,
		},
		"text": "\t\tinput.foo[x]",
	})
}

test_fail_simple_iteration_output_var_some_decl if {
	policy := ast.with_rego_v1(`allow if {
		some x
		input.foo[x]
	}`)

	r := rule.report with config.rules as allow_nesting(2)
		with input as policy
		with config.capabilities as capabilities.provided

	r == with_location({
		"col": 3,
		"file": "policy.rego",
		"row": 7,
		"end": {
			"col": 15,
			"row": 7,
		},
		"text": "\t\tinput.foo[x]",
	})
}

test_success_some_in_var_input if {
	policy := ast.with_rego_v1(`allow if {
		some x in input
		input.foo[x] == 1
	}`)

	r := rule.report with config.rules as allow_nesting(2)
		with input as policy
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_allow_nesting_zero if {
	policy := ast.with_rego_v1(`allow if {
		input.foo[_]
		input.foo[_].bar[_]
	}`)

	r := rule.report with config.rules as allow_nesting(0)
		with input as policy
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_allow_nesting_one if {
	policy := ast.with_rego_v1(`allow if {
		input.foo[_]
	}`)

	r := rule.report with config.rules as allow_nesting(1)
		with input as policy
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_allow_nesting_two if {
	policy := ast.with_rego_v1(`allow if {
		input.foo[_].bar[_]
	}`)

	r := rule.report with config.rules as allow_nesting(2)
		with input as policy
		with config.capabilities as capabilities.provided

	r == set()
}

test_fail_allow_nesting_two if {
	policy := ast.with_rego_v1(`allow if {
		input.foo[_]
	}`)

	r := rule.report with config.rules as allow_nesting(2)
		with input as policy
		with config.capabilities as capabilities.provided

	r == with_location({
		"col": 3,
		"file": "policy.rego",
		"row": 6,
		"end": {
			"col": 15,
			"row": 6,
		},
		"text": "\t\tinput.foo[_]",
	})
}

test_success_not_output_vars if {
	policy := ast.with_rego_v1(`
	x := 5

	allow if {
		y := 10
		input.foo[x].bar[y]
	}`)

	r := rule.report with config.rules as allow_nesting(2)
		with input as policy
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_output_var_to_input_var if {
	policy := ast.with_rego_v1(`allow if {
		# x is an output var here
		# iteration allowed as nesting level == 2
		input.foo[x].bar[_]
		# x is an input var here
		# iteration not allowed, but this is not iteration
		input.bar[x]
	}`)

	r := rule.report with config.rules as allow_nesting(2)
		with input as policy
		with config.capabilities as capabilities.provided

	r == set()
}

test_fail_complex_comprehension_term if {
	policy := ast.policy(`foo := [{"foo": bar} | input[bar]]`)

	r := rule.report with config.rules as allow_nesting(2)
		with input as policy
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_allow_if_subattribute if {
	policy := ast.with_rego_v1(`allow if {
		bar := input.foo[_].bar
		bar == "baz"
	}`)

	r := rule.report with config.rules as {"style": {"prefer-some-in-iteration": {
		"ignore-if-sub-attribute": true,
		"ignore-nesting-level": 5,
	}}}
		with input as policy
		with config.capabilities as capabilities.provided

	r == set()
}

test_fail_ignore_if_subattribute_disabled if {
	policy := ast.with_rego_v1(`allow if {
		bar := input.foo[_].bar
		bar == "baz"
	}`)

	r := rule.report with config.rules as {"style": {"prefer-some-in-iteration": {
		"ignore-if-sub-attribute": false,
		"ignore-nesting-level": 5,
	}}}
		with input as policy
		with config.capabilities as capabilities.provided

	r == with_location({
		"col": 10,
		"file": "policy.rego",
		"row": 6,
		"end": {
			"col": 26,
			"row": 6,
		},
		"text": "\t\tbar := input.foo[_].bar",
	})
}

test_success_allow_if_inside_array if {
	policy := ast.with_rego_v1(`allow if {
		bar := [input.foo[_] == 1]
	}`)

	r := rule.report with config.rules as allow_nesting(5)
		with input as policy
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_allow_if_inside_set if {
	r := rule.report with config.rules as allow_nesting(5)
		with input as ast.with_rego_v1(`s := {input.foo[_] == 1}`)
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_allow_if_inside_object if {
	r := rule.report with config.rules as allow_nesting(5)
		with input as ast.policy(`s := {foo: input.foo[_] == 1}`)
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_allow_if_inside_rule_head_key if {
	r := rule.report with config.rules as allow_nesting(5)
		with input as ast.with_rego_v1(`s contains input.foo[_]`)
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_allow_if_contains_check_eq if {
	r := rule.report with config.rules as allow_nesting(5)
		with input as ast.policy(`no_violation if "x" = input.foo[_]`)
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_allow_if_contains_check_equal if {
	r := rule.report with config.rules as allow_nesting(5)
		with input as ast.with_rego_v1(`no_violation if "x" == input.foo[_]`)
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_iteration_in_args if {
	r := rule.report with config.rules as allow_nesting(5)
		with input as ast.policy(`no_violation if startswith(input.foo[_], "f")`)
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_iteration_in_args_call_in_comprehension_head if {
	r := rule.report with config.rules as allow_nesting(5)
		with input as ast.policy(`r := [f(obj[k], v) | some k, v in p]`)
		with config.capabilities as capabilities.provided

	r == set()
}

test_success_top_level_iteration if {
	r := rule.report with config.rules as allow_nesting(5)
		with input as ast.policy(`r := input.foo[_]`)
		with config.capabilities as capabilities.provided

	r == set()
}

allow_nesting(i) := {"style": {"prefer-some-in-iteration": {"ignore-nesting-level": i}}}

with_location(location) := {{
	"category": "style",
	"description": "Prefer `some .. in` for iteration",
	"level": "error",
	"location": location,
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/prefer-some-in-iteration", "style"),
	}],
	"title": "prefer-some-in-iteration",
}}
