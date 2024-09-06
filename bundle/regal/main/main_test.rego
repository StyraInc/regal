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

# regal ignore:rule-length
test_main_lint if {
	ast := {
		"package": {
			"location": "1:1:cGFja2FnZQ==",
			"path": [
				{"location": "1:9:cA==", "type": "var", "value": "data"},
				{"location": "1:9:cA==", "type": "string", "value": "p"},
			],
		},
		"rules": [{
			"location": "3:1:eCA9IDE=",
			"head": {
				"name": "x",
				"ref": [{"location": "3:1:eA==", "type": "var", "value": "x"}],
				"value": {
					"value": 1,
					"location": "3:5:MQ==",
					"type": "number",
				},
				"location": "3:1:eCA9IDE=",
			},
		}],
		"regal": {
			"file": {
				"name": "p.rego",
				"lines": [
					"package p",
					"",
					"x = 1",
					"",
				],
				"abs": "/regal/p.rego",
			},
			"environment": {"path_separator": "/"},
		},
	}

	cfg := {"rules": {"style": {"use-assignment-operator": {"level": "error"}}}}

	result := main.lint with input as ast with config.merged_config as cfg

	result == {
		"aggregates": {},
		"ignore_directives": {"p.rego": {}},
		"notices": set(),
		"violations": {{
			"category": "style",
			"description": "Prefer := over = for assignment",
			"level": "error",
			"location": {
				"col": 3,
				"file": "p.rego",
				"row": 3,
				"text": "x = 1",
			},
			"related_resources": [{
				"description": "documentation",
				"ref": "https://docs.styra.com/regal/rules/style/use-assignment-operator",
			}],
			"title": "use-assignment-operator",
		}},
	}
}

test_rules_to_run_not_excluded if {
	cfg := {"rules": {"testing": {"test": {"level": "error"}}}}

	rules_to_run := main.rules_to_run with config.merged_config as cfg
		with config.for_rule as {"level": "error"}
		with input.regal.file.name as "p.rego"
		with config.excluded_file as false

	rules_to_run == {"testing": {"test"}}
}

test_notices if {
	notice := {
		"category": "idiomatic",
		"description": "here's a notice",
		"level": "notice",
		"title": "testme notice",
		"severity": "none",
	}

	notices := main.notices with main.rules_to_run as {"idiomatic": {"testme"}}
		with data.regal.rules.idiomatic.testme.notices as {notice}

	notices == {notice}
}

test_main_fail_when_input_not_object if {
	violation := {
		"category": "error",
		"title": "invalid-input",
		"description": "provided input must be a JSON AST",
	}

	report := main.report with input as []
	report == {violation}
}

test_report_custom_rule_failure if {
	report := main.report with data.custom.regal.rules as {"testing": {"testme": {"report": {{"title": "fail!"}}}}}
		with input as {"package": {}, "regal": {"file": {"name": "p.rego"}}}
		with config.excluded_file as false

	report == {{"title": "fail!"}}
}

test_aggregate_bundled_rule if {
	agg := main.aggregate with main.rules_to_run as {"foo": {"bar"}}
		with data.regal.rules as {"foo": {"bar": {"aggregate": {"baz"}}}}

	agg == {"foo/bar": {"baz"}}
}

test_aggregate_custom_rule if {
	agg := main.aggregate with data.custom.regal.rules as {"foo": {"bar": {"aggregate": {"baz"}}}}
		with config.for_rule as {"level": "error"}
		with config.excluded_file as false
		with input.regal.file.name as "p.rego"

	agg == {"foo/bar": {"baz"}}
}

test_aggregate_report_custom_rule if {
	mock_input := {
		"aggregates_internal": {"custom/test": {}},
		"regal": {"file": {"name": "p.rego"}},
	}

	mock_rules := {"custom": {"test": {"aggregate_report": {{
		"category": "custom",
		"title": "test",
	}}}}}

	report := main.aggregate_report with input as mock_input
		with data.custom.regal.rules as mock_rules

	report == {{"category": "custom", "title": "test"}}

	violations := main.lint_aggregate.violations with input as mock_input
		with data.custom.regal.rules as mock_rules

	violations == report
}
