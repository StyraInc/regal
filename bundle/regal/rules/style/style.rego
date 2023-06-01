package regal.rules.style

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	not util.is_snake_case(rule.head.name)

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some var in ast.find_vars(input.rules)
	not util.is_snake_case(var.value)

	violation := result.fail(rego.metadata.rule(), result.location(var))
}

# METADATA
# title: use-in-operator
# description: Use in to check for membership
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/use-in-operator
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some expr in eq_exprs

	expr.terms[1].type in {"array", "boolean", "object", "null", "number", "set", "string", "var"}
	expr.terms[2].type == "ref"

	last := regal.last(expr.terms[2].value)

	last.type == "var"
	startswith(last.value, "$")

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms[2].value[0]))
}

# METADATA
# title: use-in-operator
# description: Use in to check for membership
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/use-in-operator
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some expr in eq_exprs

	expr.terms[1].type == "ref"
	expr.terms[2].type in {"array", "boolean", "object", "null", "number", "set", "string", "var"}

	last := regal.last(expr.terms[1].value)

	last.type == "var"
	startswith(last.value, "$")

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms[1].value[0]))
}

eq_exprs contains expr if {
	config.for_rule({"category": "style", "title": "use-in-operator"}).level != "ignore"

	some rule in input.rules
	some expr in rule.body

	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "equal"
}

# METADATA
# title: line-length
# description: Line too long
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/line-length
# custom:
#   category: style
report contains violation if {
	cfg := config.for_rule(rego.metadata.rule())

	cfg.level != "ignore"

	some i, line in input.regal.file.lines

	line_length := count(line)
	line_length > cfg["max-line-length"]

	violation := result.fail(
		rego.metadata.rule(),
		{"location": {
			"file": input.regal.file.name,
			"row": i + 1,
			"col": line_length,
			"text": input.regal.file.lines[i],
		}},
	)
}

# METADATA
# title: use-assignment-operator
# description: Prefer := over = for assignment
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/use-assignment-operator
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	rule["default"] == true
	not rule.head.assign

	violation := result.fail(rego.metadata.rule(), result.location(rule))
}

# METADATA
# title: use-assignment-operator
# description: Prefer := over = for assignment
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/use-assignment-operator
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	rule.head.key
	rule.head.value
	not rule.head.assign

	violation := result.fail(rego.metadata.rule(), result.location(rule.head.ref[0]))
}

# For comments, OPA uses capital-cases Text and Location rather
# than text and location. As fixing this would potentially break
# things, we need to take it into consideration here.

todo_identifiers := ["todo", "TODO", "fixme", "FIXME"]

todo_pattern := sprintf(`^\s*(%s)`, [concat("|", todo_identifiers)])

# METADATA
# title: todo-comment
# description: Avoid TODO comments
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/todo-comment
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some comment in input.comments
	text := base64.decode(comment.Text)
	regex.match(todo_pattern, text)

	violation := result.fail(rego.metadata.rule(), result.location(comment))
}

# METADATA
# title: external-reference
# description: Reference to input, data or rule ref in function body
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/external-reference
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	rule.head.args

	named_args := {arg.value | some arg in rule.head.args; arg.type == "var"}
	own_vars := {v.value | some v in ast.find_vars(rule.body)}

	allowed_refs := named_args | own_vars

	some expr in rule.body

	is_array(expr.terms)

	some term in expr.terms

	term.type == "var"
	not term.value in allowed_refs

	violation := result.fail(rego.metadata.rule(), result.location(term))
}

# METADATA
# title: external-reference
# description: Reference to input, data or rule ref in function body
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/external-reference
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	rule.head.args

	named_args := {arg.value | some arg in rule.head.args; arg.type == "var"}
	own_vars := {v.value | some v in ast.find_vars(rule.body)}

	allowed_refs := named_args | own_vars

	some expr in rule.body

	is_object(expr.terms)

	terms := expr.terms.value
	terms[0].type == "var"
	not terms[0].value in allowed_refs

	violation := result.fail(rego.metadata.rule(), result.location(terms[0]))
}

# METADATA
# title: avoid-get-and-list-prefix
# description: Avoid get_ and list_ prefix for rules and functions
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/avoid-get-and-list-prefix
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	strings.any_prefix_match(rule.head.name, {"get_", "list_"})

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}

# METADATA
# title: unconditional-assignment
# description: Unconditional assignment in rule body
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/unconditional-assignment
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules

	# Single expression in rule body
	# There's going to be a few cases where more expressions
	# are in the body and it still "unconditional", like e.g
	# a `print` call.. but let's keep it simple for now
	count(rule.body) == 1

	# Var assignment in rule head
	rule.head.value.type == "var"
	rule_head_var := rule.head.value.value

	# If a `with` statement is found in body, back out, as these
	# can't be moved to the rule head
	not rule.body[0]["with"]

	# Which is an assignment (= or :=)
	terms := rule.body[0].terms
	terms[0].type == "ref"
	terms[0].value[0].type == "var"
	terms[0].value[0].value in {"eq", "assign"}

	# Of var declared in rule head
	terms[1].type == "var"
	terms[1].value == rule_head_var

	violation := result.fail(rego.metadata.rule(), result.location(terms[1]))
}
