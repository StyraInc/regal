package regal.ast

import rego.v1

keywords[row] contains keyword if {
	some rule in input.rules

	rule_text := base64.decode(rule.location.text)

	some idx, line in split(rule_text, "\n")

	col := indexof(line, " if ")
	col > 0

	row := rule.location.row + idx

	not row in _comment_row_index

	keyword := {
		"name": "if",
		"location": {
			"row": rule.location.row + idx,
			"col": col + 2,
		},
	}
}

keywords[pkg.location.row] contains keyword if {
	pkg := input["package"]

	keyword := {
		"name": "package",
		"location": {
			"row": pkg.location.row,
			"col": pkg.location.col,
		},
	}
}

keywords[imp.location.row] contains keyword if {
	some imp in input.imports

	keyword := {
		"name": "import",
		"location": {
			"row": imp.location.row,
			"col": imp.location.col,
		},
	}
}

keywords[loc.row] contains keyword if {
	some rule in input.rules

	loc := rule.head.location

	col := indexof(base64.decode(loc.text), " contains ")

	col > 0

	keyword := {
		"name": "contains",
		"location": {
			"row": loc.row,
			"col": col + 2,
		},
	}
}

keywords[value.row] contains keyword if {
	some rule in input.rules
	some expr in rule.body

	walk(expr.terms, [path, value])

	value.col
	value.row

	name := _keyword_b64s[value.text]

	parent_path := array.slice(path, 0, count(path) - 1)
	context := object.get(expr.terms, parent_path, {})

	some keyword in _determine_keywords(context, value, name)
}

_determine_keywords(_, value, name) := {keyword} if {
	name in {"in", "some"}

	keyword := {
		"name": name,
		"location": {
			"row": value.row,
			"col": value.col,
		},
	}
}

_determine_keywords(context, value, name) := keywords if {
	name == "every"

	v := object.get(context, "value", {})

	text := base64.decode(v.location.text)

	keywords := {
		{
			"name": name,
			"location": {
				"row": value.row,
				"col": value.col,
			},
		},
		{
			"name": "in",
			"location": {
				"row": value.row,
				"col": (v.location.col + count(text)) + 1,
			},
		},
	}
}

_comment_row_index contains comment.Location.row if {
	some comment in object.get(input, "comments", [])
}

_keyword_b64s := {
	"aW4=": "in",
	"c29tZQ==": "some",
	"ZXZlcnk=": "every",
}
