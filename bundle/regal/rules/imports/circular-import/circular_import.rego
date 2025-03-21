# METADATA
# description: Circular import
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/imports/circular-import
# schemas:
# - input: schema.regal.ast
package regal.rules.imports["circular-import"]

import data.regal.ast
import data.regal.result
import data.regal.util

_refs contains ref if {
	r := ast.found.refs[_][_]

	r.value[0].value == "data"

	ast.static_ref(r)

	ref := {
		"package_path": concat(".", [e.value | some e in r.value]),
		"location": object.remove(util.to_location_object(r.location), {"text"}),
	}
}

_refs contains ref if {
	some imported in ast.imports

	imported.path.value[0].value == "data"

	ref := {
		"package_path": concat(".", [e.value | some e in imported.path.value]),
		"location": object.remove(util.to_location_object(imported.path.location), {"text"}),
	}
}

# METADATA
# description: collects refs from module
aggregate contains entry if {
	count(_refs) > 0

	entry := result.aggregate(rego.metadata.chain(), {"refs": _refs})
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
aggregate_report contains violation if {
	# we need to have two or more entries in the aggregate to
	# for a circular import to be possible
	count(input.aggregate) > 1

	some g in _groups

	sorted_group := sort(g)

	location := [loc |
		some m1 in sorted_group
		some m2 in sorted_group
		some loc in _package_locations[m1][m2]
	][0]

	violation := result.fail(
		rego.metadata.chain(),
		{
			"description": sprintf("Circular import detected in: %s", [concat(", ", sort(g))]),
			"location": location,
		},
	)
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
aggregate_report contains violation if {
	# this rule tests for self dependencies
	some g in _groups

	count(g) == 1

	some pkg in g # this will the only package

	location := [e |
		some e in _package_locations[pkg][pkg]
	][0]

	violation := result.fail(
		rego.metadata.chain(),
		{
			"description": sprintf("Circular self-dependency in: %s", [pkg]),
			"location": location,
		},
	)
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
_package_locations[referenced_pkg][referencing_pkg] contains location if {
	some ag_pkg in input.aggregate

	some ref in ag_pkg.aggregate_data.refs

	referenced_pkg := ref.package_path
	referencing_pkg := sprintf("data.%s", [concat(".", ag_pkg.aggregate_source.package_path)])
	ref_loc := util.to_location_object(ref.location)

	location := {
		"file": ag_pkg.aggregate_source.file,
		"row": ref_loc.row,
		"col": ref_loc.col,
	}
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
_import_graph[pkg] contains edge if {
	some ag_pkg in input.aggregate

	pkg := sprintf("data.%s", [concat(".", ag_pkg.aggregate_source.package_path)])

	some pkg_ref in ag_pkg.aggregate_data.refs

	edge := pkg_ref.package_path
}

_reachable_index[pkg] := reachable if {
	some pkg, _ in _import_graph

	reachable := graph.reachable(_import_graph, {pkg})
}

_self_reachable contains pkg if {
	some pkg, _ in _import_graph

	pkg in _reachable_index[pkg]
}

_groups contains group if {
	some pkg in _self_reachable

	# only consider packages that have edges to other packages,
	# even if only to themselves
	_import_graph[pkg] != {}

	reachable := graph.reachable(_import_graph, {pkg})

	group := {m |
		some m in reachable
		pkg in _reachable_index[m]
	}
}
