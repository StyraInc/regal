package regal.rules.style_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style
import data.regal.rules.style.common_test.report

test_fail_use_in_operator_string_lhs if {
	r := report(`allow {
	"admin" == input.user.roles[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 13, "file": "policy.rego", "row": 9, "text": "\t\"admin\" == input.user.roles[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_number_lhs if {
	r := report(`allow {
	1 == input.lucky_numbers[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 7, "file": "policy.rego", "row": 9, "text": "\t1 == input.lucky_numbers[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_array_lhs if {
	r := report(`allow {
	[1] == input.arrays[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 9, "file": "policy.rego", "row": 9, "text": "\t[1] == input.arrays[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_boolean_lhs if {
	r := report(`allow {
	true == input.booleans[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 10, "file": "policy.rego", "row": 9, "text": "\ttrue == input.booleans[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_object_lhs if {
	r := report(`allow {
	{"x": "y"} == input.objects[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 16, "file": "policy.rego", "row": 9, "text": "\t{\"x\": \"y\"} == input.objects[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_null_lhs if {
	r := report(`allow {
	null == input.objects[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 10, "file": "policy.rego", "row": 9, "text": "\tnull == input.objects[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_set_lhs if {
	r := report(`allow {
	{"foo"} == input.objects[_]
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 13, "file": "policy.rego", "row": 9, "text": "\t{\"foo\"} == input.objects[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_var_lhs if {
	report(`allow {
	admin == input.user.roles[_]
	}`) == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 11, "file": "policy.rego", "row": 9, "text": "\tadmin == input.user.roles[_]"},
		"level": "error",
	}}
}

test_fail_use_in_operator_string_rhs if {
	report(`allow {
	input.user.roles[_] == "admin"
	}`) == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 2, "file": "policy.rego", "row": 9, "text": "\tinput.user.roles[_] == \"admin\""},
		"level": "error",
	}}
}

test_fail_use_in_operator_var_rhs if {
	r := report(`allow {
		input.user.roles[_] == admin
	}`)
	r == {{
		"category": "style",
		"description": "Use in to check for membership",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-in-operator", "style"),
		}],
		"title": "use-in-operator",
		"location": {"col": 3, "file": "policy.rego", "row": 9, "text": "\t\tinput.user.roles[_] == admin"},
		"level": "error",
	}}
}

test_success_refs_both_sides if {
	report(`allow { required_roles[_] == input.user.roles[_] }`) == set()
}

test_success_uses_in_operator if {
	report(`allow { "admin" in input.user.roles[_] }`) == set()
}
