package regal.result

import future.keywords.if
import future.keywords.in

import data.regal.config

# METADATA
# description: |
#  The result.aggregate function works similarly to `result.fail`, but instead of producing
#  a violation returns an entry to be aggregated for later evaluation. This is useful in
#  aggregate rules (and only in aggregate rules) as it provides a uniform format for
#  aggregate data entries. Example return value:
#
#  {
#      "rule": {
#          "category": "testing",
#          "title": "aggregation",
#      },
#      "aggregate_source": {
#          "file": "policy.rego",
#          "package_path": ["a", "b", "c"],
#      },
#      "aggregate_data": {
#          "foo": "bar",
#          "baz": [1, 2, 3],
#      },
#  }
#
aggregate(chain, aggregate_data) := entry if {
	is_array(chain)

	some link in chain
	link.annotations.scope == "package"

	[category, title] := _category_title_from_path(link.path)

	entry := {
		"rule": {
			"category": category,
			"title": title,
		},
		"aggregate_source": {
			"file": input.regal.file.name,
			"package_path": [part.value |
				some i, part in input["package"].path
				i > 0
			],
		},
		"aggregate_data": aggregate_data,
	}
}

_category_title_from_path(path) := [category, title] if ["regal", "rules", category, title] = path

_category_title_from_path(path) := [category, title] if ["custom", "regal", "rules", category, title] = path

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

	annotation := object.union(link.annotations, {
		"custom": {"category": category},
		"title": title,
	})

	violation := _fail_annotated_custom(annotation, details)
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
	is_object(metadata)
	with_location := object.union(metadata, details)
	category := with_location.custom.category
	with_category := object.union(with_location, {
		"category": category,
		"level": config.rule_level(config.for_rule(category, metadata.title)),
	})

	without_custom_and_scope := object.remove(with_category, ["custom", "scope", "schemas"])
	related_resources := resource_urls(without_custom_and_scope.related_resources, category)

	violation := json.patch(
		without_custom_and_scope,
		[{"op": "replace", "path": "/related_resources", "value": related_resources}],
	)
}

_fail_annotated_custom(metadata, details) := violation if {
	is_object(metadata)
	with_location := object.union(metadata, details)
	category := with_location.custom.category
	with_category := object.union(with_location, {
		"category": category,
		"level": config.rule_level(config.for_rule(category, metadata.title)),
	})

	violation := object.remove(with_category, ["custom", "scope", "schemas"])
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

location(x) := with_text(x[0].location) if is_array(x)

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
