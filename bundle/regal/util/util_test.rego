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
