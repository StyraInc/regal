package regal.util_test

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

test_point_in_range if {
	util.point_in_range([1, 2], [[0, 0], [1, 10]]) == true
	util.point_in_range([0, 3], [[0, 1], [0, 4]]) == true
	util.point_in_range([0, 0], [[0, 0], [0, 2]]) == true
	util.point_in_range([0, 2], [[0, 0], [0, 2]]) == true
	util.point_in_range([6, 6], [[5, 10], [7, 3]]) == true

	util.point_in_range([0, 0], [[0, 1], [1, 10]]) == false
	util.point_in_range([0, 3], [[0, 1], [0, 2]]) == false
	util.point_in_range([9, 3], [[0, 1], [0, 2]]) == false
}
