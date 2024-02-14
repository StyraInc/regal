# METADATA
# description: Rule name repeats package
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/style/rule-name-repeats-package
# schemas:
# - input: schema.regal.ast
package regal.rules.style["rule-name-repeats-package"]

import rego.v1

import data.regal.result

titleize(str) := upper(str) if count(str) == 1

titleize(str) := result if {
	chrs := regex.split(``, str)
	count(chrs) > 1

	result := concat(
		"",
		array.concat([upper(chrs[0])], array.slice(chrs, 1, count(chrs))),
	)
}

package_path := input["package"].path

package_path_components := [component.value |
	some component in package_path
	component.type == "string"
]

num_package_path_components := count(package_path_components)

possible_path_component_combinations contains combination if {
	some end in numbers.range(1, num_package_path_components)

	combination := array.slice(
		package_path_components,
		num_package_path_components - end,
		num_package_path_components,
	)
}

possible_offending_prefixes contains prefix if {
	some combination in possible_path_component_combinations

	prefix := concat("_", combination)
}

possible_offending_prefixes contains prefix if {
	some combination in possible_path_component_combinations

	count(combination) > 1

	formatted_combination := array.concat(
		[combination[0]],
		[w |
			some word in array.slice(combination, 1, count(combination))
			w := titleize(word)
		],
	)

	prefix := concat("", formatted_combination)
}

report contains violation if {
	some rule in input.rules

	strings.any_prefix_match(rule.head.name, possible_offending_prefixes)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
