package regal.ast_test

import rego.v1

import data.regal.ast

test_keywords_package if {
	policy := `package policy`

	kwds := ast.keywords with input as regal.parse_module("p.rego", policy)

	count(kwds) == 1 # lines with keywords

	_keyword_on_row(
		kwds,
		1,
		{
			"name": "package",
			"location": {"row": 1, "col": 1},
		},
	)
}

test_keywords_import if {
	policy := `package policy

import rego.v1`

	kwds := ast.keywords with input as regal.parse_module("p.rego", policy)

	count(kwds) == 2 # lines with keywords

	_keyword_on_row(
		kwds,
		3,
		{
			"name": "import",
			"location": {"row": 3, "col": 1},
		},
	)
}

test_keywords_if if {
	policy := `package policy

import rego.v1

allow if {
	# if things
	true
}
`

	kwds := ast.keywords with input as regal.parse_module("p.rego", policy)

	count(kwds) == 3 # lines with keywords

	_keyword_on_row(
		kwds,
		5,
		{
			"name": "if",
			"location": {"row": 5, "col": 7},
		},
	)
}

test_keywords_some_in if {
	policy := `package policy

import rego.v1

allow if {
	some e in [1,2,3]
}`

	kwds := ast.keywords with input as regal.parse_module("p.rego", policy)

	count(kwds) == 4 # lines with keywords

	_keyword_on_row(
		kwds,
		6,
		{"name": "some", "location": {"row": 6, "col": 2}},
	)

	_keyword_on_row(
		kwds,
		6,
		{"name": "in", "location": {"row": 6, "col": 9}},
	)
}

test_keywords_contains if {
	policy := `package policy

import rego.v1

messages contains "hello" if {
	1 == 1
}`

	kwds := ast.keywords with input as regal.parse_module("p.rego", policy)

	count(kwds) == 3 # lines with keywords

	_keyword_on_row(
		kwds,
		5,
		{"name": "contains", "location": {"row": 5, "col": 10}},
	)

	_keyword_on_row(
		kwds,
		5,
		{"name": "if", "location": {"row": 5, "col": 27}},
	)
}

test_keywords_every if {
	policy := `package policy

import rego.v1

allow if {
	every k in [1,2,3] {
		k == "foo"
		v == 1
	}
}`

	kwds := ast.keywords with input as regal.parse_module("p.rego", policy)

	count(kwds) == 4 # lines with keywords

	_keyword_on_row(
		kwds,
		6,
		{"name": "every", "location": {"row": 6, "col": 2}},
	)

	_keyword_on_row(
		kwds,
		6,
		{"name": "in", "location": {"row": 6, "col": 10}},
	)
}

_keyword_on_row(kwds, row, keyword) if {
	some kwd in object.get(kwds, row, {})

	kwd.name == keyword.name
	kwd.location.row == keyword.location.row
	kwd.location.col == keyword.location.col
}
