package regal.config_test

import future.keywords.if

import data.regal.config

# map[pattern: string]map[filename: string]excluded: bool
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
		subcases := cases[pattern]
		res := {file: res1 |
			exp := subcases[file]
			act := config.exclude(pattern, file)
			exp != act
			res1 := {"exp": exp, "act": act}
		}
		count(res) > 0
	}

	count(not_exp) == 0
}

rules_config_error := {"rules": {"test": {"test-case": {"level": "error"}}}}

rules_config_ignore_delta := {"rules": {"test": {"test-case": {"ignore": {"files": ["p.rego"]}}}}}

config_ignore := {"ignore": {"files": ["p.rego"]}}

test_excluded_file_default if {
	e := config.excluded_file("test", "test-case", "p.rego") with data.eval.params as params
		with config.merged_config as rules_config_error

	e == false
}

test_excluded_file_with_ignore if {
	c := object.union(rules_config_error, rules_config_ignore_delta)
	e := config.excluded_file("test", "test-case", "p.rego") with data.eval.params as params
		with config.merged_config as c
	e == true
}

test_excluded_file_config if {
	e := config.excluded_file("test", "test-case", "p.rego") with config.merged_config as config_ignore
	e == true
}

test_excluded_file_cli_flag if {
	e := config.excluded_file("test", "test-case", "p.rego") with data.eval.params as object.union(
		params,
		{"ignore_files": ["p.rego"]},
	)
	e == true
}

test_excluded_file_cli_overrides_config if {
	e := config.excluded_file("test", "test-case", "p.rego") with config.merged_config as config_ignore
		with data.eval.params as object.union(params, {"ignore_files": [""]})
	e == false
}
