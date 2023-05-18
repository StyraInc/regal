package regal.result

import future.keywords.if
import future.keywords.in

import data.regal.config

fail(metadata, details) := violation if {
	with_location := object.union(metadata, details)
	category := with_location.custom.category
	with_category := object.union(with_location, {
		"category": category,
		"level": config.rule_level(config.for_rule(with_location)),
	})

	without_custom_and_scope := object.remove(with_category, ["custom", "scope"])
	related_resources := resource_urls(without_custom_and_scope.related_resources, category)

	violation := object.union(
		object.remove(without_custom_and_scope, ["related_resources"]),
		{"related_resources": related_resources},
	)
}

resource_urls(related_resources, category) := [r |
	some item in related_resources
	r := object.union(object.remove(item, ["ref"]), {"ref": config.docs.resolve_url(item.ref, category)})
]

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
