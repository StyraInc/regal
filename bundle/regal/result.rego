package regal.result

import rego.v1

import data.regal.config
import data.regal.util

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
			# regal ignore:external-reference
			"file": input.regal.file.name,
			"package_path": [part.value |
				# regal ignore:external-reference
				some i, part in input["package"].path
				i > 0
			],
		},
		"aggregate_data": aggregate_data,
	}
}

_category_title_from_path(path) := [category, title] if ["regal", "rules", category, title] = path

_category_title_from_path(path) := [category, title] if ["custom", "regal", "rules", category, title] = path

# METADATA
# description: |
#   helper function to call when building the "return value" for the `report` in any linter rule —
#   recommendation being that both built-in rules and custom rules use this in favor of building the
#   result by hand
# scope: document

# METADATA
# description: provided rules, i.e. regal.rules.category.title
fail(metadata, details) := violation if {
	is_array(metadata) # from rego.metadata.chain()

	some link in metadata
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

# METADATA
# description: custom rules, i.e. custom.regal.rules.category.title
fail(metadata, details) := violation if {
	is_array(metadata) # from rego.metadata.chain()

	some link in metadata
	link.annotations.scope == "package"

	some category, title
	["custom", "regal", "rules", category, title] = link.path

	annotation := object.union(link.annotations, {
		"custom": {"category": category},
		"title": title,
	})

	violation := _fail_annotated_custom(annotation, details)
}

# METADATA
# description: fallback case
fail(metadata, details) := _fail_annotated(metadata, details)

# METADATA
# description: |
#   creates a notice object, i.e. one used to inform users of things like a rule getting
#   ignored because the set capabilities does not include a dependency (like a built-in function)
#   needed by the rule
notice(metadata) := result if {
	is_array(metadata)
	rule_meta := metadata[0]
	annotations := rule_meta.annotations

	some category, title
	["regal", "rules", category, title, "notices"] = rule_meta.path

	result := {
		"category": category,
		"description": annotations.description,
		"level": "notice",
		"title": title,
		"severity": annotations.custom.severity,
	}
}

_related_resources(annotations, _, _) := annotations.related_resources

_related_resources(annotations, category, title) := rr if {
	not annotations.related_resources
	rr := [{
		"description": "documentation",
		# regal ignore:external-reference
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

resource_urls(related_resources, category) := [r |
	some item in related_resources
	r := object.union(object.remove(item, ["ref"]), {"ref": config.docs.resolve_url(item.ref, category)})
]

_with_text(loc_obj) := loc if {
	loc := {"location": object.union(loc_obj, {
		"file": input.regal.file.name, # regal ignore:external-reference
		"text": input.regal.file.lines[loc_obj.row - 1], # regal ignore:external-reference
	})}

	loc_obj.row
} else := {"location": loc_obj}

location(x) := _with_text(util.to_location_object(x.location))

location(x) := _with_text(util.to_location_object(x[0].location)) if is_array(x)

# Special case for rule refs, where location is currently only assigned to the value
# In this case, we'll just assume that the column is 1, as that's the most likely place
# See: https://github.com/open-policy-agent/opa/issues/5790
location(x) := _with_text(loc_obj) if {
	not x.location
	count(x.ref) == 1
	x.ref[0].type == "var"
	loc_obj := object.union(util.to_location_object(x.value.location), {"col": 1})
}

location(x) := {} if {
	not x.location
	not x.Location
	count(x.ref) != 1
}

# METADATA
# description: |
#   similar to `location` but includes an `end` attribute where `end.row` is the same
#   value as x.location.row, and `end.col` is the start col + the length of the text
#   original attribute value base64 decoded from x.location
ranged_location_from_text(x) := end if {
	otext := base64.decode(util.to_location_object(x.location).text)
	lines := split(otext, "\n")

	loc := location(x)
	end := object.union(loc, {"location": {"end": {
		# note the use of the _original_ location text here, as loc.location.text
		# will be the entire line where the violation occurred
		"row": (loc.location.row + count(lines)) - 1,
		"col": loc.location.col + count(regal.last(lines)),
	}}})
}

# METADATA
# description: creates a location where x is the start, and y is the end (calculated from `text`)
ranged_location_between(x, y) := object.union(
	location(x),
	{"location": {"end": ranged_location_from_text(y).location.end}},
)
