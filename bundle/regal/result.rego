package regal.result

import future.keywords.if
import future.keywords.in

import data.regal.config

# Provided rules, i.e. regal.rules.category.title
fail(chain, details) := violation if {
	is_array(chain) # from rego.metadata.chain()

	some link in chain
	link.annotations.scope == "package"

	some category, title
	["regal", "rules", category, title] = link.path

	# NOTE: consider allowing annotation to override any derived values?
	annotation := object.union(link.annotations, {
		"custom": {"category": category},
		"title": title,
		"related_resources": _related_resources(link.annotations, category, title),
	})

	violation := _fail_annotated(annotation, details)
}

# Custom rules, i.e. custom.regal.rules.category.title
fail(chain, details) := violation if {
	is_array(chain) # from rego.metadata.chain()

	some link in chain
	link.annotations.scope == "package"

	some category, title
	["custom", "regal", "rules", category, title] = link.path

	# NOTE: consider allowing annotation to override any derived values?
	annotation := object.union(link.annotations, {
		"custom": {"category": category},
		"title": title,
		"related_resources": _related_resources(link.annotations, category, title),
	})

	violation := _fail_annotated(annotation, details)
}

_related_resources(annotations, _, _) := annotations.related_resources

_related_resources(annotations, category, title) := rr if {
	not annotations.related_resources
	rr := [{
		"description": "documentation",
		"ref": sprintf("%s/%s/%s", [config.docs.base_url, category, title]),
	}]
}

_fail_annotated(metadata, details) := violation if {
	is_object(metadata) # from rego.metadata.rule()
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

fail(metadata, details) := _fail_annotated(metadata, details)

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
