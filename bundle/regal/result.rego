package regal.result

fail(metadata, details) := violation {
	with_location := object.union(metadata, details)
	with_category := object.union(with_location, {"category": with_location.custom.category})

	violation := object.remove(with_category, ["custom", "scope"])
}

location(x) := {"location": x.location} {
	x.location
}

# Special case for comments, where this typo sadly is hard to change at this point
location(x) := {"location": x.Location} {
	x.Location
}

# Special case for rule refs, where location is currently only assigned to the value
# In this case, we'll just assume that the column is 1, as that's the most likely place
# See: https://github.com/open-policy-agent/opa/issues/5790
location(x) := {"location": location} {
	not x.location
	count(x.ref) == 1
	x.ref[0].type == "var"
	location := object.union(x.value.location, {"col": 1})
}

location(x) := {} {
	not x.location
	not x.Location
	count(x.ref) != 1
}
