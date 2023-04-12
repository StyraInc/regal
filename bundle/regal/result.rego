package regal.result

import future.keywords.if

fail(metadata, details) := violation if {
	with_location := object.union(metadata, details)
	with_category := object.union(with_location, {"category": with_location.custom.category})

	violation := object.remove(with_category, ["custom", "scope"])
}

with_text(location) := {"location": object.union(location, {"text": input.regal.file.lines[location.row - 1]})} if {
	location.row
} else := {"location": location}

location(x) := with_text(x.location) if x.location

# Special case for comments, where this typo sadly is hard to change at this point
location(x) := with_text(x.Location) if x.Location

# Special case for rule refs, where location is currently only assigned to the value
# In this case, we'll just assume that the column is 1, as that's the most likely place
# See: https://github.com/open-policy-agent/opa/issues/5790
location(x) := with_text(location) if {
	not x.location
	count(x.ref) == 1
	x.ref[0].type == "var"
	location := object.union(x.value.location, {"col": 1})
}

location(x) := {} if {
	not x.location
	not x.Location
	count(x.ref) != 1
}
