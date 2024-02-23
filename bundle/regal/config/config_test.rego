package regal.config_test

import rego.v1

import data.regal.config

rules_config := {"rules": {"test": {"test-case": {
	"level": "ignore",
	"important_setting": 42,
}}}}

params := {
	"disable_all": false,
	"disable_category": [],
	"disable": [],
	"enable_all": false,
	"enable_category": [],
	"enable": [],
}

# disable all

test_disable_all_no_config if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"disable_all": true})

	c == {"level": "ignore"}
}

test_disable_all_with_config if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"disable_all": true})
		with config.merged_config as rules_config

	c == {"level": "ignore", "important_setting": 42}
}

test_disable_all_with_category_override if {
	p := object.union(params, {"disable_all": true, "enable_category": ["test"]})
	c := config.for_rule("test", "test-case") with data.eval.params as p
		with config.merged_config as rules_config

	c == {"level": "error", "important_setting": 42}
}

test_disable_all_with_rule_override if {
	p := object.union(params, {"disable_all": true, "enable": ["test-case"]})
	c := config.for_rule("test", "test-case") with data.eval.params as p
		with config.merged_config as rules_config

	c == {"level": "error", "important_setting": 42}
}

# disable category

test_disable_category_no_config if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"disable_category": ["test"]})

	c == {"level": "ignore"}
}

test_disable_category_with_config if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"disable_category": ["test"]})
		with config.merged_config as rules_config

	c == {"level": "ignore", "important_setting": 42}
}

test_disable_category_with_rule_override if {
	p := object.union(params, {"disable_category": ["test"], "enable": ["test-case"]})
	c := config.for_rule("test", "test-case") with data.eval.params as p
		with config.merged_config as rules_config

	c == {"level": "error", "important_setting": 42}
}

# disable rule

test_disable_single_rule if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"disable": ["test-case"]})

	c == {"level": "ignore"}
}

test_disable_single_rule_with_config if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"disable": ["test-case"]})
		with config.merged_config as rules_config

	c == {"level": "ignore", "important_setting": 42}
}

# enable all

test_enable_all_no_config if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"enable_all": true})

	c == {"level": "error"}
}

test_enable_all_with_config if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"enable_all": true})
		with config.merged_config as rules_config

	c == {"level": "error", "important_setting": 42}
}

test_enable_all_with_category_override if {
	p := object.union(params, {"enable_all": true, "disable_category": ["test"]})
	c := config.for_rule("test", "test-case") with data.eval.params as p
		with config.merged_config as rules_config

	c == {"level": "ignore", "important_setting": 42}
}

test_enable_all_with_rule_override if {
	p := object.union(params, {"enable_all": true, "disable": ["test-case"]})
	c := config.for_rule("test", "test-case") with data.eval.params as p
		with config.merged_config as rules_config

	c == {"level": "ignore", "important_setting": 42}
}

# disable category

test_enable_category_no_config if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"enable_category": ["test"]})

	c == {"level": "error"}
}

test_enable_category_with_config if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"enable_category": ["test"]})
		with config.merged_config as rules_config

	c == {"level": "error", "important_setting": 42}
}

test_enable_category_with_rule_override if {
	p := object.union(params, {"enable_category": ["test"], "disable": ["test-case"]})
	c := config.for_rule("test", "test-case") with data.eval.params as p
		with config.merged_config as rules_config

	c == {"level": "ignore", "important_setting": 42}
}

# enable rule

test_enable_single_rule if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"enable": ["test-case"]})

	c == {"level": "error"}
}

test_enable_single_rule_with_config if {
	c := config.for_rule("test", "test-case") with data.eval.params as object.union(params, {"enable": ["test-case"]})
		with config.merged_config as rules_config

	c == {"level": "error", "important_setting": 42}
}

test_all_rules_are_in_provided_configuration if {
	missing_config := {title |
		some category, title
		data.regal.rules[category][title]
		not endswith(title, "_test")
		not config.provided.rules[category][title]
	}

	count(missing_config) == 0
}

test_all_configured_rules_exist if {
	go_rules := {"opa-fmt"}

	missing_rules := {title |
		some category, title
		config.provided.rules[category][title]
		not data.regal.rules[category][title]
	}

	count(missing_rules - go_rules) == 0
}
