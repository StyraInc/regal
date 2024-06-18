# METADATA
# description: various rules and functions related to location and position
package regal.lsp.completion.location

import rego.v1

import data.regal.ast

# METADATA
# description: best-effort helper to determine if the current line is in a rule body
# scope: document
in_rule_body(line) if contains(line, " if ")

in_rule_body(line) if contains(line, " contains ")

in_rule_body(line) if contains(line, " else ")

in_rule_body(line) if contains(line, "= ")

in_rule_body(line) if regex.match(`^\s+`, line)

# METADATA
# description: converts OPA location to LSP position
to_position(location) := {
	"line": location.row - 1,
	"character": location.col - 1,
}

# METADATA
# description: |
#   estimate where the location "ends" based on its text attribute,
#   both line and column
end_location_estimate(location) := end if {
	lines := split(base64.decode(location.text), "\n")
	end := {
		"row": (location.row + count(lines)) - 1,
		"col": count(regal.last(lines)),
	}
}

# METADATA
# description: |
#   find and return rule at provided location
#   undefined if provided location is not within the range of a rule
find_rule(rules, location) := [rule |
	some i, rule in rules
	end_location := end_location_estimate(rule.location)
	location.row >= rule.location.row
	location.row <= end_location.row
][0]

# METADATA
# description: |
#   find local variables (declared via function arguments, some/every
#   declarations or assignment) at the given location
find_locals(rules, location) := ast.find_names_in_local_scope(find_rule(rules, location), location)

# METADATA
# description: |
#   return the range for a word object (as return by `word_at`)
word_range(word, position) := {
	"start": {
		"line": position.line,
		"character": position.character - word.offset_before,
	},
	"end": {
		"line": position.line,
		"character": position.character + word.offset_after,
	},
}

# METADATA
# description: |
#   find word at column in line, and return its text, and the offset
#   from the position (before and after)
word_at(line, col) := word if {
	text_before := substring(line, 0, col - 1)
	word_before := _to_string(regex.find_n(`[a-zA-Z_]+$`, text_before, 1))

	text_after := substring(line, col - 1, count(line))
	word_after := _to_string(regex.find_n(`^[a-zA-Z_]+`, text_after, 1))

	word := {
		"offset_before": count(word_before),
		"offset_after": count(word_after),
		"text": sprintf("%s%s", [word_before, word_after]),
	}
}

_to_string(arr) := "" if count(arr) == 0

_to_string(arr) := arr[0] if count(arr) > 0
