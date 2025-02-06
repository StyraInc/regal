# METADATA
# description: Non loop expression in loop
package regal.rules.performance["non-loop-expression"]

import data.regal.ast
import data.regal.result
import data.regal.util

_exprs[sprintf("%d", [rule_index])][row] contains expr if {
	some rule_index
	expr := input.rules[rule_index].body[_]
	row := to_number(substring(expr.location, 0, indexof(expr.location, ":")))
}

_exprs[sprintf("%d", [rule_index])][row] contains expr if {
	some rule_index
	expr := input.rules[rule_index].body[_].terms.body[_]

	row := to_number(substring(expr.location, 0, indexof(expr.location, ":")))
}

_loop_start_points[rule_index][loc.row] contains var if {
	some rule_index
	var := ast.found.vars[rule_index].assign[_]
	term := _exprs[rule_index][_][_][_]

	last := regal.last(regal.last(term).value)
	last.type == "var"
	startswith(last.value, "$")

	loc := util.to_location_object(var.location)
	loc.row == util.to_location_object(term[0].location).row
	# no need to ignore vars here in comprehensions, since we are only looking
	# for top level wildcards in the final term.
}

_loop_start_points[rule_index][loc.row] contains var if {
	some rule_index
	some context in ["some", "somein", "every"]
	var := ast.found.vars[rule_index][context][_]

	loc := util.to_location_object(var.location)

	# ignore vars in comprehensions
	comps := object.get(ast.found.comprehensions, rule_index, set())
	every comp in comps {
		comp_loc := util.to_location_object(comp.location)
		range := [[comp_loc.row, comp_loc.col], [comp_loc.end.row, comp_loc.end.col]]
		not util.point_in_range([loc.row, loc.col], range)
		not util.point_in_range([loc.end.row, loc.end.col], range)
	}
}

_loop_start_points[rule_index][row] contains var if {
	some rule_index
	call := ast.function_calls[rule_index][_]

	call.name == "walk"
	call.args[1].type == "array"

	some var in ast.find_term_vars(call.args[1].value)
	row := to_number(substring(var.location, 0, indexof(var.location, ":")))
}

_assignment_index[rule_index][var.value] contains row if {
	some rule_index, row
	some var in _loop_start_points[rule_index][row]
}

_assignment_index[rule_index][var.value] contains loc.row if {
	some rule_index
	some var in ast.found.vars[rule_index].assign
	loc := util.to_location_object(var.location)

	# ignore vars in comprehensions
	comps := object.get(ast.found.comprehensions, rule_index, set())
	every comp in comps {
		comp_loc := util.to_location_object(comp.location)
		range := [[comp_loc.row, comp_loc.col], [comp_loc.end.row, comp_loc.end.col]]
		not util.point_in_range([loc.row, loc.col], range)
		not util.point_in_range([loc.end.row, loc.end.col], range)
	}
}

report contains violation if {
	some rule_index, sps in _loop_start_points
	first_loop_row := min(object.keys(sps))

	some row
	some expr in _exprs[rule_index][row]
	row > first_loop_row

	# if there are any term vars used in the expression, then they must have been
	# declared after the first loop
	every term_var in ast.find_term_vars(expr) {
		term_var_rows := object.get(_assignment_index, [rule_index, term_var.value], {0})
		min(term_var_rows) < first_loop_row
	}

	violation := result.fail(rego.metadata.chain(), result.location(expr))
}
