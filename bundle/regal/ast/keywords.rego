package regal.ast

import data.regal.util

# METADATA
# description: collects keywords from input module by the line that they appear on
# scope: document

# METADATA
# description: collects the `if` keyword. this isn't present in the AST, so we'll simply scan the input lines
keywords[row] contains keyword if {
	some idx, line in input.regal.file.lines

	col := indexof(line, " if ")
	col > 0

	row := idx + 1

	not row in _comment_row_index

	keyword := {
		"name": "if",
		"location": {
			"row": idx + 1,
			"col": col + 2,
		},
	}
}

# METADATA
# description: collects the `package` keyword
keywords[loc.row] contains keyword if {
	pkg := input["package"]
	loc := util.to_location_object(pkg.location)

	keyword := {
		"name": "package",
		"location": {
			"row": loc.row,
			"col": loc.col,
		},
	}
}

# METADATA
# description: collects the `import` keyword
keywords[loc.row] contains keyword if {
	some imp in input.imports

	loc := util.to_location_object(imp.location)

	keyword := {
		"name": "import",
		"location": {
			"row": loc.row,
			"col": loc.col,
		},
	}
}

# METADATA
# description: collects the `contains` keyword
keywords[loc.row] contains keyword if {
	some rule in input.rules

	loc := util.to_location_object(rule.head.location)
	col := indexof(loc.text, " contains ")

	col > 0

	keyword := {
		"name": "contains",
		"location": {
			"row": loc.row,
			"col": col + 2,
		},
	}
}

# METADATA
# description: collects the `some`, `every` and `in` keywords
keywords[keyword.location.row] contains keyword if {
	walk(input.rules, [_, value])

	some keyword in _keywords_with_location(value)
}

_keywords_with_location(value) := keywords if {
	value.terms.symbols

	location := util.to_location_object(value.terms.location)
	keywords := array.concat([{"name": "some", "location": location}], _in_on_row(location))
}

_keywords_with_location(value) := keywords if {
	value.domain

	location := util.to_location_object(value.location)
	keywords := array.concat([{"name": "every", "location": location}], _in_on_row(location))
}

_in_on_row(location) := [keyword |
	in_col := indexof(input.regal.file.lines[location.row - 1], " in ")
	keyword := {
		"name": "in",
		"location": {
			"row": location.row,
			"col": in_col + 2,
			"end": {
				"row": location.row,
				"col": in_col + 4,
			},
			"text": "in",
		},
	}
]

_comment_row_index contains util.to_location_object(comment.location).row if some comment in input.comments

_keywords := {
	"in",
	"some",
	"every",
}
