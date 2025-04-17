package regal.config_test

import data.regal.config

params(override) := object.union(
	{
		"disable_all": false,
		"disable_category": [],
		"disable": [],
		"enable_all": false,
		"enable_category": [],
		"enable": [],
	},
	override,
)

test_disable_all_no_config if {
	c := config.level_for_rule("test", "test-case") with data.eval.params as params({"disable_all": true})

	c == "ignore"
}

test_enable_all_no_config if {
	c := config.level_for_rule("test", "test-case") with data.eval.params as params({"enable_all": true})

	c == "error"
}

test_config[name] if {
	some name, [override, want_level] in {
		"disable_all_with_config": [{"disable_all": true}, "ignore"],
		"disable_all_with_category_override": [{"disable_all": true, "enable_category": ["test"]}, "error"],
		"disable_all_with_rule_override": [{"disable_all": true, "enable": ["test-case"]}, "error"],
		"disable_category_no_config": [{"disable_category": ["test"]}, "ignore"],
		"disable_category_with_config": [{"disable_category": ["test"]}, "ignore"],
		"disable_category_with_rule_override": [{"disable_category": ["test"], "enable": ["test-case"]}, "error"],
		"disable_single_rule": [{"disable": ["test-case"]}, "ignore"],
		"disable_single_rule_with_config": [{"disable": ["test-case"]}, "ignore"],
		"enable_all_with_config": [{"enable_all": true}, "error"],
		"enable_all_with_category_override": [{"enable_all": true, "disable_category": ["test"]}, "ignore"],
		"enable_all_with_rule_override": [{"enable_all": true, "disable": ["test-case"]}, "ignore"],
		"enable_category_no_config": [{"enable_category": ["test"]}, "error"],
		"enable_category_with_config": [{"enable_category": ["test"]}, "error"],
		"enable_category_with_rule_override": [{"enable_category": ["test"], "disable": ["test-case"]}, "ignore"],
		"enable_single_rule": [{"enable": ["test-case"]}, "error"],
		"enable_single_rule_with_config": [{"enable": ["test-case"]}, "error"],
	}

	l := config.level_for_rule("test", "test-case") with data.eval.params as params(override)

	l == want_level

	c := config.for_rule("test", "test-case") with config.merged_config as {"rules": {"test": {"test-case": {
		"level": "ignore",
		"important_setting": 42,
	}}}}

	c == {"level": "ignore", "important_setting": 42}
}

test_all_rules_are_in_provided_configuration if {
	missing_config := {title |
		some category, title
		data.regal.rules[category][title].report
		not endswith(title, "_test")
		not config.provided.rules[category][title]
	}

	count(missing_config) == 0
}

test_all_configured_rules_exist if {
	missing_rules := {title |
		some category, title
		config.provided.rules[category][title]
		not data.regal.rules[category][title]
	}

	count(missing_rules) == 0
}

test_path_prefix if {
	config.path_prefix == ""
	config.path_prefix == "foo" with data.internal.path_prefix as "foo"
}
