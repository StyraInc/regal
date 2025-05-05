# METADATA
# description: Rule name repeats package
# related_resources:
#   - description: documentation
#     ref: https://docs.styra.com/regal/rules/style/rule-name-repeats-package
package regal.rules.style["rule-name-repeats-package"]

import data.regal.ast
import data.regal.result
import data.regal.util

# METADATA
# description: reports any location where a rule name repeats the package name
report contains violation if {
	some rule in input.rules

	strings.any_prefix_match(ast.ref_static_to_string(rule.head.ref), _possible_offending_prefixes)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head.ref[0]))
}

_titleize(str) := upper(str) if count(str) == 1

_titleize(str) := concat("", array.concat([upper(chrs[0])], array.slice(chrs, 1, count(chrs)))) if {
	chrs := regex.split(``, str)
	count(chrs) > 1
}

_num_package_path_components := count(ast.package_path)

_possible_path_component_combinations contains combination if {
	some end in numbers.range(1, _num_package_path_components)

	combination := array.slice(
		ast.package_path,
		_num_package_path_components - end,
		_num_package_path_components,
	)
}

_possible_offending_prefixes contains concat("_", combination) if {
	some combination in _possible_path_component_combinations
}

_possible_offending_prefixes contains concat("", formatted_combination) if {
	some combination in _possible_path_component_combinations

	count(combination) > 1

	formatted_combination := array.concat([combination[0]], [w |
		some word in util.rest(combination)
		w := _titleize(word)
	])
}
