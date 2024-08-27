# METADATA
# description: various utility functions for linter policies
package regal.util

import rego.v1

# METADATA
# description: returns true if string is snake_case formatted
is_snake_case(str) if str == lower(str)

# METADATA
# description: |
#   returns a set of sets containing all indices of duplicates in the array,
#   so e.g. [1, 1, 2, 3, 3, 3] would return {{0, 1}, {3, 4, 5}} and so on
find_duplicates(arr) := {indices |
	some i, x in arr

	indices := {j |
		some j, y in arr
		x == y
	}

	count(indices) > 1
}

# METADATA
# description: returns true if array has duplicates of item
has_duplicates(arr, item) if count([x |
	some x in arr
	x == item
]) > 1

# METADATA
# description: |
#   returns an array of arrays built from all parts of the provided path array,
#   so e.g. [1, 2, 3] would return [[1], [1, 2], [1, 2, 3]]
all_paths(path) := [array.slice(path, 0, len) | some len in numbers.range(1, count(path))]

# METADATA
# description: attempts to turn any key in provided object into numeric form
keys_to_numbers(obj) := {num: v |
	some k, v in obj
	num := to_number(k)
}

to_location_object(loc) := {"row": to_number(row), "col": to_number(col), "text": text} if {
	is_string(loc)
	[row, col, text] := split(loc, ":")
}

to_location_object(loc) := loc if is_object(loc)

# METADATA
# description: short-hand helper to prepare values for pretty-printing
json_pretty(value) := json.marshal_with_options(value, {
	"indent": "  ",
	"pretty": true,
})
