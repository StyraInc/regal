package regal.util_test

import rego.v1

import data.regal.util

test_find_duplicates if {
	util.find_duplicates([1, 1, 2, 3, 3, 3]) == {{0, 1}, {3, 4, 5}}
}

test_json_pretty if {
	# oh, the things you do for test coverage
	util.json_pretty({"x": [1, 2, 3]}) == `{
  "x": [
    1,
    2,
    3
  ]
}`
}

test_rest if {
	util.rest([1, 2, 3]) == [2, 3]
	util.rest([1]) == []
	util.rest([]) == []
}

test_to_location_object if {
	loc := util.to_location_object("3:1:5:2") with input.regal.file.lines as [
		"package p",
		"",
		"allow if {",
		"\ttrue",
		"}",
	]

	loc == {
		"row": 3,
		"col": 1,
		"end": {
			"row": 5,
			"col": 2,
		},
		"text": "allow if {\n\ttrue\n}",
	}
}
