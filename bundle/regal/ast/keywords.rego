package regal.ast

import rego.v1

import data.regal.util

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

keywords[loc.row] contains keyword if {
	some rule in input.rules

	loc := util.to_location_object(rule.head.location)
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

keywords[loc.row] contains keyword if {
	expr := exprs[_][_]

	walk(expr.terms, [path, value])

	regal.last(path) == "location"

	loc := util.to_location_object(value)
	name := _keyword_b64s[loc.text]

	parent_path := array.slice(path, 0, count(path) - 1)
	context := object.get(expr.terms, parent_path, {})

	some keyword in _determine_keywords(context, loc, name)
}

keywords[loc.row] contains keyword if {
	some rule in input.rules
	rule.head.assign

	walk(rule.head.value, [path, value])

	regal.last(path) == "location"

	loc := util.to_location_object(value)

	name := _keyword_b64s[loc.text]

	parent_path := array.slice(path, 0, count(path) - 1)
	context := object.get(rule.head.value, parent_path, {})

	some keyword in _determine_keywords(context, loc, name)
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

_determine_keywords(context, value, "every") := keywords if {
	ctx_loc := util.to_location_object(context.value.location)
	text := base64.decode(ctx_loc.text)

	keywords := {
		{
			"name": "every",
			"location": {
				"row": value.row,
				"col": value.col,
			},
		},
		{
			"name": "in",
			"location": {
				"row": value.row,
				"col": (ctx_loc.col + count(text)) + 1,
			},
		},
	}
}

_comment_row_index contains util.to_location_object(comment.location).row if some comment in input.comments

_keyword_b64s := {
	"aW4=": "in",
	"c29tZQ==": "some",
	"ZXZlcnk=": "every",
}
