package regal.util

import rego.v1

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
# description: |
#   returns an array of arrays built from all parts of the provided path array,
#   so e.g. [1, 2, 3] would return [[1], [1, 2], [1, 2, 3]]
all_paths(path) := [array.slice(path, 0, len) | some len in numbers.range(1, count(path))]
