package regal.rules.style["prefer-some-in-iteration_test"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.rules.style["prefer-some-in-iteration"] as rule

test_fail_simple_iteration if {
	policy := ast.policy(`allow {
		input.foo[_] == "bar"
	}`)

	r := rule.report with config.for_rule as allow_nesting(2) with input as policy
	r == with_location({"col": 3, "file": "policy.rego", "row": 4, "text": "\t\tinput.foo[_] == \"bar\""})
}

test_fail_simple_iteration_output_var if {
	policy := ast.policy(`allow {
		input.foo[x] == 1
	}`)

	r := rule.report with config.for_rule as allow_nesting(2) with input as policy
	r == with_location({"col": 3, "file": "policy.rego", "row": 4, "text": "\t\tinput.foo[x] == 1"})
}

test_fail_simple_iteration_output_var_some_decl if {
	policy := ast.policy(`allow {
		some x
		input.foo[x] == 1
	}`)

	r := rule.report with config.for_rule as allow_nesting(2) with input as policy
	r == with_location({"col": 3, "file": "policy.rego", "row": 5, "text": "\t\tinput.foo[x] == 1"})
}

test_success_some_in_var_input if {
	policy := ast.with_rego_v1(`allow if {
		some x in input
		input.foo[x] == 1
	}`)

	r := rule.report with config.for_rule as allow_nesting(2) with input as policy
	r == set()
}

test_success_allow_nesting_zero if {
	policy := ast.policy(`allow {
		input.foo[_] == 1
		input.foo[_].bar[_] == 2
	}`)

	r := rule.report with config.for_rule as allow_nesting(0) with input as policy
	r == set()
}

test_success_allow_nesting_one if {
	policy := ast.policy(`allow {
		input.foo[_] == 2
	}`)

	r := rule.report with config.for_rule as allow_nesting(1) with input as policy
	r == set()
}

test_success_allow_nesting_two if {
	policy := ast.policy(`allow {
		input.foo[_].bar[_] == 2
	}`)

	r := rule.report with config.for_rule as allow_nesting(2) with input as policy
	r == set()
}

test_fail_allow_nesting_two if {
	policy := ast.policy(`allow {
		input.foo[_] == 2
	}`)

	r := rule.report with config.for_rule as allow_nesting(2) with input as policy
	r == with_location({"col": 3, "file": "policy.rego", "row": 4, "text": "\t\tinput.foo[_] == 2"})
}

test_success_not_output_vars if {
	policy := ast.policy(`
	x := 5

	allow {
		y := 10
		input.foo[x].bar[y] == 2
	}`)

	r := rule.report with config.for_rule as allow_nesting(2) with input as policy
	r == set()
}

test_success_output_var_to_input_var if {
	policy := ast.policy(`allow {
		# x is an output var here
		# iteration allowed as nesting level == 2
		input.foo[x].bar[_]
		# x is an input var here
		# iteration not allowed, but this is not iteration
		input.bar[x]
	}`)

	r := rule.report with config.for_rule as allow_nesting(2) with input as policy
	r == set()
}

test_success_complex_comprehension_term if {
	policy := ast.policy(`

	foo := [{"foo": bar} | input[bar]]
	`)

	r := rule.report with config.for_rule as allow_nesting(2) with input as policy
	r == set()
}

test_success_allow_if_subattribute if {
	policy := ast.policy(`allow {
		bar := input.foo[_].bar
		bar == "baz"
	}`)

	r := rule.report with config.for_rule as {
		"level": "error",
		"ignore-if-sub-attribute": true,
		"ignore-nesting-level": 5,
	}
		with input as policy
	r == set()
}

test_fail_ignore_if_subattribute_disabled if {
	policy := ast.policy(`allow {
		bar := input.foo[_].bar
		bar == "baz"
	}`)

	r := rule.report with config.for_rule as {
		"level": "error",
		"ignore-if-sub-attribute": false,
		"ignore-nesting-level": 5,
	}
		with input as policy
	r == with_location({"col": 10, "file": "policy.rego", "row": 4, "text": "\t\tbar := input.foo[_].bar"})
}

allow_nesting(i) := {
	"level": "error",
	"ignore-nesting-level": i,
}

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
