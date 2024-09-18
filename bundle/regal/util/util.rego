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

# METADATA
# description: convert location string to location object
# scope: document
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

# METADATA
# description: returns all elements of arr after the first
rest(arr) := array.slice(arr, 1, count(arr))

# METADATA
# description: converts x to set if array, returns x if set
# scope: document
to_set(x) := x if is_set(x)

to_set(x) := {y | some y in x} if not is_set(x)

# METADATA
# description: true if s1 and s2 has any intersecting items
intersects(s1, s2) if count(intersection({s1, s2})) > 0

# METADATA
# description: returns the item contained in a single-item set
single_set_item(s) := item if {
	count(s) == 1

	some item in s
}

# METADATA
# description: returns any item of a set
any_set_item(s) := [x | some x in s][0] # this is convoluted.. but can't think of a better way

# METADATA
# description: returns last index of item, or undefined (*not* -1) if missing
last_indexof(arr, item) := regal.last([i | some i, x in arr; x == item])
