# METADATA
# description: various utility functions for linter policies
package regal.util

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
to_location_object(loc) := {
	"row": row,
	"col": col,
	"text": text,
	"end": {
		"row": end_row,
		"col": end_col,
	},
} if {
	is_string(loc)

	[r, c, er, ec] := split(loc, ":")

	row := to_number(r)
	col := to_number(c)
	end_row := to_number(er)
	end_col := to_number(ec)

	text := _location_to_text(row, col, end_row, end_col)
}

to_location_object(loc) := loc if is_object(loc)

_location_to_text(row, col, end_row, end_col) := substring(
	input.regal.file.lines[row - 1],
	col - 1,
	end_col - col,
) if {
	row == end_row
}

_location_to_text(row, col, end_row, end_col) := text if {
	row != end_row

	lines := array.slice(input.regal.file.lines, row - 1, end_row)
	text := concat("\n", [new |
		len := count(lines) - 1

		some i, line in lines

		new := _cut_col(i, len, line, col, end_col)
	])
}

_cut_col(0, 1, line, col, end_col) := substring(line, col - 1, end_col - 1)
_cut_col(0, len, line, _, _) := line if len > 1

_cut_col(i, len, line, _, end_col) := substring(line, 0, end_col) if {
	i == len
} else := line if {
	i > 0
}

# METADATA
# scope: document
# description: returns true if point is within range of row,col range
default point_in_range(_, _) := false

point_in_range(p, range) if {
	p[0] >= range[0][0]
	p[0] <= range[1][0]
	p[1] >= range[0][1]
	p[1] <= range[1][1]
}

point_in_range(p, range) if {
	p[0] > range[0][0]
	p[0] < range[1][0]
}

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
# description: converts x to array if set, returns x if array
# scope: document
to_array(x) := x if is_array(x)
to_array(x) := [y | some y in x] if not is_array(x)

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

# METADATA
# description: |
#   returns the longest common 'prefix' sequence found in coll (set or array of arrays)
#   e.g. [[1, 2, 3, 4], [1, 2, 4], [1, 2, 5]] would return [1, 2]
#   if any of the passed collections are empty, the result is an empty array
longest_prefix(coll) := [] if {
	[] in coll
} else := prefix if {
	arr := to_array(coll)
	end := min([count(seq) | some seq in arr]) - 1
	rng := numbers.range(0, end)
	eqn := max([n |
		some n in rng

		first := arr[0][n]
		every sub in arr {
			sub[n] == first
		}
	])

	prefix := array.slice(arr[0], 0, eqn + 1)
}
