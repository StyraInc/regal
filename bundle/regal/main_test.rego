package regal.main_test

test_main_basic_input_success {
	report := data.regal.main.report with input as regal.parse_module("p.rego", `package p`)
	report == set()
}

test_main_multiple_failues {
	policy := `package p

	# both camel case and unification operator
	default camelCase = "yes"
	`
	report := data.regal.main.report with input as regal.parse_module("p.rego", policy)
		with data.regal.config.for_rule as {"level": "error"}

	count(report) == 2
}

test_main_ignore_directive_failure {
	policy := `package p

	# regal ignore:todo-comment
	camelCase := "yes"
	`
	report := data.regal.main.report with input as regal.parse_module("p.rego", policy)
		with data.regal.config.for_rule as {"level": "error"}

	count(report) == 1
}

test_main_ignore_directive_success {
	policy := `package p

	# regal ignore:prefer-snake-case
	camelCase := "yes"
	`
	report := data.regal.main.report with input as regal.parse_module("p.rego", policy)
		with data.regal.config.for_rule as {"level": "error"}

	count(report) == 0
}

test_main_ignore_directive_multiple_success {
	policy := `package p

	# regal ignore:prefer-snake-case,use-assignment-operator
	default camelCase = "yes"
	`
	report := data.regal.main.report with input as regal.parse_module("p.rego", policy)
		with data.regal.config.for_rule as {"level": "error"}

	count(report) == 0
}

test_main_ignore_directive_multiple_mixed_success {
	policy := `package p

	# regal ignore:prefer-snake-case,todo-comment
	default camelCase = "yes"
	`
	report := data.regal.main.report with input as regal.parse_module("p.rego", policy)
		with data.regal.config.for_rule as {"level": "error"}

	count(report) == 1
}
