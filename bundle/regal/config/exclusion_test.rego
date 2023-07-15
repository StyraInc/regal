package regal.config_test

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
}

test_all_cases_are_as_expected {
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

rule_specific_config := {"ignore": ["p.rego"]}

config_ignore := {"ignore": ["p.rego"]}

test_excluded_file_default {
	e := config.excluded_file({}, "p.rego")
	e == false
}

test_excluded_file_with_ignore {
	e := config.excluded_file(rule_specific_config, "p.rego")
	e == true
}

test_excluded_file_config {
	e := config.excluded_file({}, "p.rego") with config.merged_config as config_ignore
	e == true
}

test_excluded_file_cli_flag {
	e := config.excluded_file({}, "p.rego") with data.eval.params as object.union(params, {"ignore": ["p.rego"]})
	e == true
}

test_excluded_file_cli_overrides_config {
	e := config.excluded_file({}, "p.rego") with config.merged_config as config_ignore
		with data.eval.params as object.union(params, {"ignore": [""]})
	e == false
}
