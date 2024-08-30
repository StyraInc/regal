package regal.main_test

import rego.v1

import data.regal.config
import data.regal.main

test_basic_input_success if {
	report := main.report with input as regal.parse_module("p.rego", `package p`)
	report == set()
}

test_multiple_failures if {
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

test_expect_failure if {
	policy := `package p

	camelCase := "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "error"}}}}

	count(report) == 1
}

test_ignore_rule_config if {
	policy := `package p

	camelCase := "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "ignore"}}}}

	count(report) == 0
}

test_ignore_directive_failure if {
	policy := `package p

	# regal ignore:todo-comment
	camelCase := "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "error"}}}}

	count(report) == 1
}

test_ignore_directive_success if {
	policy := `package p

	# regal ignore:prefer-snake-case
	camelCase := "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "error"}}}}

	count(report) == 0
}

test_ignore_directive_success_same_line if {
	policy := `package p

	camelCase := "yes" # regal ignore:prefer-snake-case
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "error"}}}}

	count(report) == 0
}

test_ignore_directive_success_same_line_trailing_directive if {
	policy := `package p

	camelCase := "yes" # camelCase is nice! # regal ignore:prefer-snake-case
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "error"}}}}

	count(report) == 0
}

test_ignore_directive_success_same_line_todo_comment if {
	policy := `package p

	camelCase := "yes" # TODO! camelCase isn't nice! # regal ignore:todo-comment
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"todo-comment": {"level": "error"}}}}

	count(report) == 0
}

test_ignore_directive_multiple_success if {
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

test_ignore_directive_multiple_mixed_success if {
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

test_ignore_directive_collected_in_aggregate_rule if {
	module := regal.parse_module("p.rego", `package p

	import rego.v1

	# regal ignore:unresolved-import
	import data.unresolved
	`)
	lint := main.lint with input as module

	lint.ignore_directives == {"p.rego": {6: ["unresolved-import"]}}
}

test_ignore_directive_enforced_in_aggregate_rule if {
	report_without_ignore_directives := main.aggregate_report with input as {
		"aggregates_internal": {"imports/unresolved-import": []},
		"regal": {"file": {"name": "p.rego"}},
		"ignore_directives": {},
	}
		with config.merged_config as {"rules": {"imports": {"unresolved-import": {"level": "error"}}}}
		with data.regal.rules.imports["unresolved-import"].aggregate_report as {{
			"category": "imports",
			"level": "error",
			"location": {"col": 1, "file": "p.rego", "row": 6, "text": "import data.provider.parameters"},
			"title": "unresolved-import",
		}}

	count(report_without_ignore_directives) == 1

	report_with_ignore_directives := main.aggregate_report with input as {
		"aggregates_internal": {"imports/unresolved-import": []},
		"regal": {"file": {"name": "p.rego"}},
		"ignore_directives": {"p.rego": {"6": ["unresolved-import"]}},
	}
		with config.merged_config as {"rules": {"imports": {"unresolved-import": {"level": "error"}}}}
		with data.regal.rules.imports["unresolved-import"].aggregate_report as {{
			"category": "imports",
			"level": "error",
			"location": {"col": 1, "file": "p.rego", "row": 6, "text": "import data.provider.parameters"},
			"title": "unresolved-import",
		}}

	count(report_with_ignore_directives) == 0
}

test_exclude_files_rule_config if {
	policy := `package p

	camelCase := "yes"
	`
	cfg := {"rules": {"style": {"prefer-snake-case": {"level": "error", "ignore": {"files": ["p.rego"]}}}}}
	report := main.report with input as regal.parse_module("p.rego", policy) with config.merged_config as cfg

	count(report) == 0
}

test_force_exclude_file_eval_param if {
	policy := `package p

	camelCase := "yes"
	`
	report := main.report with input as regal.parse_module("p.rego", policy)
		with config.merged_config as {"rules": {"style": {"prefer-snake-case": {"level": "error"}}}}
		with data.eval.params.ignore_files as ["p.rego"]

	count(report) == 0
}

test_force_exclude_file_config if {
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
