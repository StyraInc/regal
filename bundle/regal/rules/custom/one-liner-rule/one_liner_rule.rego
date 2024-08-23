# METADATA
# description: Rule body could be made a one-liner
package regal.rules.custom["one-liner-rule"]

import rego.v1

import data.regal.ast
import data.regal.capabilities
import data.regal.config
import data.regal.result
import data.regal.util

cfg := config.for_rule("custom", "one-liner-rule")

# METADATA
# description: Missing capability for keyword `if`
# custom:
#   severity: warning
notices contains result.notice(rego.metadata.chain()) if not capabilities.has_if

report contains violation if {
	# No need to traverse rules here if we're not importing `if`
	ast.imports_keyword(input.imports, "if")

	# Note: this covers both rules and functions, which is what we want here
	some rule in input.rules

	# Bail out of rules with else for now. It is possible that they can be made
	# one-liners, but they'll often be longer than the preferred line length
	# We can come back to this later, but for now let's just make this an
	# exception documented for this rule
	not rule["else"]

	# Single expression in body required for one-liner
	count(rule.body) == 1

	# Note that this will give us the text representation of the whole rule,
	# which we'll need as the "if" is only visible here ¯\_(ツ)_/¯
	text := base64.decode(util.to_location_object(rule.location).text)
	lines := [line |
		some s in split(text, "\n")
		line := trim_space(s)
	]

	# Technically, the `if` could be on another line, but who would do that?
	regex.match(`\s+if`, lines[0])
	rule_body_brackets(lines)

	# ideally we'd take style preference into account but for now assume tab == 4 spaces
	# then just add the sum of the line counts minus the removed '{' character
	# redundant parens added by `opa fmt` :/
	((4 + count(lines[0])) + count(lines[1])) - 1 < max_line_length

	not comment_in_body(rule, object.get(input, "comments", []), lines)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

# K&R style
rule_body_brackets(lines) if endswith(lines[0], "{")

# Allman style
rule_body_brackets(lines) if {
	not endswith(lines[0], "{")
	startswith(lines[1], "{")
}

comment_in_body(rule, comments, lines) if {
	rule_location := util.to_location_object(rule.location)

	some comment in comments

	comment_location := util.to_location_object(comment.location)

	comment_location.row > rule_location.row
	comment_location.row < rule_location.row + count(lines)
}

default max_line_length := 120

max_line_length := cfg["max-line-length"]
