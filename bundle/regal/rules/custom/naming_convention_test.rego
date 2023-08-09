package regal.rules.custom["naming-convention_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.custom["naming-convention"] as rule

test_fail_package_name_does_not_match_pattern if {
	cfg := {
		"level": "error",
		"conventions": [{"targets": ["package"], "pattern": `^foo\.bar\..+$`}],
	}
	r := rule.report with input as regal.parse_module("policy.rego", "package foo.bar") with config.for_rule as cfg
	r == {{
		"category": "custom",
		# regal ignore:line-length
		"description": "Naming convention violation: package name \"foo.bar\" does not match pattern \"^foo\\\\.bar\\\\..+$\"",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 1, "text": "package foo.bar"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/naming-convention", "custom"),
		}],
		"title": "naming-convention",
	}}
}

test_success_package_name_matches_pattern if {
	cfg := {
		"level": "error",
		"conventions": [{"targets": ["package"], "pattern": `^foo\.bar$`}],
	}
	r := rule.report with input as regal.parse_module("policy.rego", "package foo.bar") with config.for_rule as cfg
	r == set()
}

test_fail_rule_name_does_not_match_pattern if {
	cfg := {
		"level": "error",
		"conventions": [{"targets": ["rule"], "pattern": "^[a-z]+$"}],
	}
	r := rule.report with input as ast.policy(`FOO := true`) with config.for_rule as cfg
	r == {{
		"category": "custom",
		"description": "Naming convention violation: rule name \"FOO\" does not match pattern \"^[a-z]+$\"",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "FOO := true"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/naming-convention", "custom"),
		}],
		"title": "naming-convention",
	}}
}

test_success_rule_name_matches_pattern if {
	cfg := {
		"level": "error",
		"conventions": [{"targets": ["rule"], "pattern": "^[a-z]+$"}],
	}
	r := rule.report with input as ast.policy(`foo := true`) with config.for_rule as cfg
	r == set()
}

test_fail_function_name_does_not_match_pattern if {
	cfg := {
		"level": "error",
		"conventions": [{"targets": ["function"], "pattern": "^[a-z]+$"}],
	}
	r := rule.report with input as ast.policy(`fooBar(_) := true`) with config.for_rule as cfg
	r == {{
		"category": "custom",
		"description": "Naming convention violation: function name \"fooBar\" does not match pattern \"^[a-z]+$\"",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "fooBar(_) := true"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/naming-convention", "custom"),
		}],
		"title": "naming-convention",
	}}
}

test_success_function_name_matches_pattern if {
	cfg := {
		"level": "error",
		"conventions": [{"targets": ["function"], "pattern": "^[a-z_]+$"}],
	}
	r := rule.report with input as ast.policy(`foo_bar(_) := true`) with config.for_rule as cfg
	r == set()
}

test_fail_var_name_does_not_match_pattern if {
	policy := ast.policy(`
	allow {
		fooBar := true
		fooBar == true
	}
	`)
	cfg := {
		"level": "error",
		"conventions": [{"targets": ["variable"], "pattern": "^[a-z_]+$"}],
	}
	r := rule.report with input as policy with config.for_rule as cfg
	r == {{
		"category": "custom",
		"description": "Naming convention violation: variable name \"fooBar\" does not match pattern \"^[a-z_]+$\"",
		"level": "error",
		"location": {"col": 3, "file": "policy.rego", "row": 5, "text": "\t\tfooBar := true"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/naming-convention", "custom"),
		}],
		"title": "naming-convention",
	}}
}

test_success_var_name_matches_pattern if {
	policy := ast.policy(`
	allow {
		some foo_bar
		input[foo_bar]
		foo_bar == "works"
	}
	`)
	cfg := {
		"level": "error",
		"conventions": [{"targets": ["variable"], "pattern": "^[a-z_]+$"}],
	}
	r := rule.report with input as policy with config.for_rule as cfg
	r == set()
}

test_fail_multiple_conventions if {
	policy := regal.parse_module("policy.rego", `package foo.bar

	foo := true

	bar {
		fooBar := true
		fooBar == true
	}
	`)
	cfg := {
		"level": "error",
		"conventions": [
			{"targets": ["package"], "pattern": `^acmecorp\.[a-z_\.]+$`},
			{"targets": ["rule", "variable"], "pattern": "^bar$|^foo_bar$"},
		],
	}
	r := rule.report with input as policy with config.for_rule as cfg
	r == {
		{
			"category": "custom",
			# regal ignore:line-length
			"description": "Naming convention violation: package name \"foo.bar\" does not match pattern \"^acmecorp\\\\.[a-z_\\\\.]+$\"",
			"level": "error",
			"location": {"col": 1, "file": "policy.rego", "row": 1, "text": "package foo.bar"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/naming-convention", "custom"),
			}],
			"title": "naming-convention",
		},
		{
			"category": "custom",
			"description": "Naming convention violation: rule name \"foo\" does not match pattern \"^bar$|^foo_bar$\"",
			"level": "error",
			"location": {"col": 2, "file": "policy.rego", "row": 3, "text": "\tfoo := true"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/naming-convention", "custom"),
			}],
			"title": "naming-convention",
		},
		{
			"category": "custom",
			"description": "Naming convention violation: variable name \"fooBar\" does not match pattern \"^bar$|^foo_bar$\"",
			"level": "error",
			"location": {"col": 3, "file": "policy.rego", "row": 6, "text": "\t\tfooBar := true"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/naming-convention", "custom"),
			}],
			"title": "naming-convention",
		},
	}
}
