package regal.rules.idiomatic["directory-package-mismatch_test"]

import rego.v1

import data.regal.config

import data.regal.rules.idiomatic["directory-package-mismatch"] as rule

test_success_directory_structure_package_path_match if {
	module := regal.parse_module("foo/bar/baz/p.rego", "package foo.bar.baz")
	r := rule.report with input as module with config.for_rule as default_config

	r == set()
}

test_fail_directory_structure_package_path_mismatch if {
	module := regal.parse_module("foo/bar/baz/p.rego", "package foo.baz.bar")
	r := rule.report with input as module with config.for_rule as default_config

	r == with_location({"col": 9, "file": "foo/bar/baz/p.rego", "row": 1, "text": "package foo.baz.bar"})
}

test_success_directory_structure_package_path_match_longer_directory_path if {
	module := regal.parse_module("system/directories/foo/bar/baz/p.rego", "package foo.bar.baz")
	r := rule.report with input as module with config.for_rule as default_config

	r == set()
}

# note that this is not considered a violation but a _notice_ of severity warning
# see corresponding test below
test_success_directory_structure_package_path_match_shorter_directory_path if {
	module := regal.parse_module("bar/baz/p.rego", "package foo.bar.baz")
	r := rule.report with input as module with config.for_rule as default_config

	r == set()
}

test_notice_severity_warning_when_directory_path_shorter_than_package_path if {
	module := regal.parse_module("bar/baz/p.rego", "package foo.bar.baz")
	r := rule.notices with input as module with config.for_rule as default_config

	r == {{
		"category": "idiomatic",
		"description": "package 'foo.bar.baz' has more parts than provided directory path 'bar/baz'",
		"level": "notice",
		"severity": "warning",
		"title": "directory-package-mismatch",
	}}
}

test_notice_severity_none_when_no_path_likely_single_file_provided if {
	module := regal.parse_module("p.rego", "package p")
	r := rule.notices with input as module with config.for_rule as default_config

	r == {{
		"category": "idiomatic",
		"description": "provided file has no directory components in its path... try linting a directory",
		"level": "notice",
		"severity": "none",
		"title": "directory-package-mismatch",
	}}
}

with_location(location) := {{
	"category": "idiomatic",
	"description": "Directory structure should mirror package",
	"level": "error",
	"location": location,
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/directory-package-mismatch", "idiomatic"),
	}],
	"title": "directory-package-mismatch",
}}

default_config := {
	"level": "error",
	"exclude-test-suffix": true,
}
