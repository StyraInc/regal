package regal.config_test

import rego.v1

import data.regal.config

cases := {
	"p.rego": {
		"p.rego": true,
		"dir/p.rego": true,
		"dir1/dir2/p.rego": true,
		"q.rego": false,
	},
	"**/p.rego": {
		"p.rego": true,
		"dir/p.rego": true,
		"dir1/dir2/p.rego": true,
		"q.rego": false,
	},
	"dir/**/p.rego": {
		"dir/sub/p.rego": true,
		"dir2/sub/p.rego": false,
		"dir/p.rego": true,
		"dir2/p.rego": false,
		"dir/sub/sub1/p.rego": true,
		"dir/sub/q.rego": false,
	},
	"*_test.rego": {
		"p_test.rego": true,
		"q_test.rego": true,
		"dir/p_test.rego": true,
		"p.rego": false,
		"ptest.rego": false,
	},
	"outer/inner/": {
		"outer/inner/file": true,
		"outer/inner": false,
	},
	"/p.rego": {
		"p.rego": true,
		"dir/p.rego": false,
		"p.rego/file": true,
	},
}

test_all_cases_are_as_expected if {
	not_exp := {pattern: res |
		some pattern, subcases in cases
		res := {file |
			some file, exp in subcases
			act := config._exclude(pattern, file)
			exp != act
		}
		count(res) > 0
	}

	count(not_exp) == 0
}

rules_config_error := {"rules": {"test": {"test-case": {"level": "error"}}}}

rules_config_ignore_delta := {"rules": {"test": {"test-case": {"ignore": {"files": ["p.rego"]}}}}}

config_ignore := {"ignore": {"files": ["p.rego"]}}

test_excluded_file_default if {
	not config.excluded_file("test", "test-case", "p.rego") with data.eval.params as params
		with config.merged_config as rules_config_error
}

test_excluded_file_with_ignore if {
	c := object.union(rules_config_error, rules_config_ignore_delta)

	config.excluded_file("test", "test-case", "p.rego") with data.eval.params as params
		with config.merged_config as c
}

test_excluded_file_config if {
	config.excluded_file("test", "test-case", "p.rego") with config.merged_config as config_ignore
}

test_excluded_file_cli_flag if {
	config.excluded_file("test", "test-case", "p.rego") with data.eval.params as object.union(
		params,
		{"ignore_files": ["p.rego"]},
	)
}

test_excluded_file_cli_overrides_config if {
	not config.excluded_file("test", "test-case", "p.rego") with config.merged_config as config_ignore
		with data.eval.params as object.union(params, {"ignore_files": [""]})
}

test_trailing_slash if {
	config._trailing_slash("foo/**/bar") == {"foo/**/bar", "foo/**/bar/**"}
	config._trailing_slash("foo") == {"foo", "foo/**"}
	config._trailing_slash("foo/**") == {"foo/**"}
}
