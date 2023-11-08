package regal.main_test

import future.keywords.if

import data.regal.config
import data.regal.main

test_main_basic_input_success if {
	report := main.report with input as regal.parse_module("p.rego", `package p`)
	report == set()
}

test_main_multiple_failures if {
	policy := `package p

	# both camel case and unification operator
	default camelCase = "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {
			"prefer-snake-case": {"level": "error"},
			"use-assignment-operator": {"level": "error"},
		}}}

	count(report) == 2
}

test_main_expect_failure if {
	policy := `package p

	camelCase := "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "error"}}}}

	count(report) == 1
}

test_main_ignore_rule_config if {
	policy := `package p

	camelCase := "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "ignore"}}}}

	count(report) == 0
}

test_main_ignore_directive_failure if {
	policy := `package p

	# regal ignore:todo-comment
	camelCase := "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "error"}}}}

	count(report) == 1
}

test_main_ignore_directive_success if {
	policy := `package p

	# regal ignore:prefer-snake-case
	camelCase := "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "error"}}}}

	count(report) == 0
}

test_main_ignore_directive_success_same_line if {
	policy := `package p

	camelCase := "yes" # regal ignore:prefer-snake-case
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "error"}}}}

	count(report) == 0
}

test_main_ignore_directive_multiple_success if {
	policy := `package p

	# regal ignore:prefer-snake-case,use-assignment-operator
	default camelCase = "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {
			"prefer-snake-case": {"level": "error"},
			"use-assignment-operator": {"level": "error"},
		}}}

	count(report) == 0
}

test_main_ignore_directive_multiple_mixed_success if {
	policy := `package p

	# regal ignore:prefer-snake-case,todo-comment
	default camelCase = "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {
			"prefer-snake-case": {"level": "error"},
			"use-assignment-operator": {"level": "error"},
		}}}

	count(report) == 1
}

test_main_exclude_files_rule_config if {
	policy := `package p

	camelCase := "yes"
	`
	cfg := {"rules": {"style": {"prefer-snake-case": {"level": "error", "ignore": {"files": ["p.rego"]}}}}}
	report := main.report with input as regal.parse_module("p.rego", policy) with config.merged_config as cfg

	count(report) == 0
}

test_main_force_exclude_file_eval_param if {
	policy := `package p

	camelCase := "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "error"}}}}
		with data.eval.params.ignore_files as ["p.rego"]

	count(report) == 0
}

test_main_force_exclude_file_config if {
	policy := `package p

	camelCase := "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {
			"rules": {"style": {"prefer-snake-case": {"level": "error"}}},
			"ignore": {"files": ["p.rego"]},
		}

	count(report) == 0
}
