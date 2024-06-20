package regal.lsp.completion.providers.rulerefs

import rego.v1

import data.regal.ast
import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

ref_is_internal(ref) if contains(ref, "._")

default determine_ref_prefix(_) := ""

determine_ref_prefix(word) := word if {
	word != ":="
}

position := location.to_position(input.regal.context.location)

line := input.regal.file.lines[position.line]

last_word := regal.last(regex.split(`\s+`, trim_space(line)))

workspace_rule_refs contains ref if {
	some refs in data.workspace.defined_refs
	some ref in refs
}

parsed_current_file := data.workspace.parsed[input.regal.file.uri]

current_file_package := concat(".", [segment.value |
	some segment in parsed_current_file["package"].path
])

current_file_imports contains ref if {
	some imp in parsed_current_file.imports

	ref := ast.ref_to_string(imp.path.value)
}

current_package_refs contains ref if {
	some ref in workspace_rule_refs

	startswith(ref, current_file_package)
}

imported_package_refs contains ref if {
	some ref in workspace_rule_refs
	some pkg_ref in current_file_imports

	not ref_is_internal(ref)

	startswith(ref, pkg_ref)
}

other_package_refs contains ref if {
	some ref in workspace_rule_refs

	not ref in imported_package_refs
	not ref in current_package_refs
}

# from the current package
rule_ref_suggestions contains pkg_ref if {
	some ref in current_package_refs

	pkg_ref := trim_prefix(ref, sprintf("%s.", [current_file_package]))
}

# from imported packages
rule_ref_suggestions contains pkg_ref if {
	some ref in imported_package_refs
	some imported_package in current_file_imports

	startswith(ref, imported_package)

	parts := split(imported_package, ".")
	prefix := concat(".", array.slice(parts, 0, count(parts) - 1))
	pkg_ref := trim_prefix(ref, sprintf("%s.", [prefix]))
}

# from any other package
rule_ref_suggestions contains ref if {
	some ref in other_package_refs

	not ref_is_internal(ref)
}

# also suggest the unimported packages themselves
# e.g. data.foo.rule will also generate data.foo as a suggestion
rule_ref_suggestions contains pkg if {
	some ref in other_package_refs

	not ref_is_internal(ref)

	parts := split(ref, ".")
	pkg := concat(".", array.slice(parts, 0, count(parts) - 1))
}

matching_rule_ref_suggestions contains ref if {
	line != ""
	location.in_rule_body(line)

	prefix := determine_ref_prefix(last_word)

	some ref in rule_ref_suggestions

	startswith(ref, prefix)

	# \W is used here to match ( in the case of func() := ..., as well as the space in the case of rule := ...
	first_word := regex.split(`\W+`, trim_space(line))[0]

	# this is to avoid suggesting a recursive rule, e.g. rule := rule, or func() := func()
	ref != first_word
}

grouped_refs[size] contains ref if {
	some ref in matching_rule_ref_suggestions

	size := count(indexof_n(ref, "."))
}

items := [item |
	some _, group in grouped_refs
	some ref in sort(group)

	item := {
		"label": ref,
		"kind": kind.variable,
		"detail": "rule ref",
		"textEdit": {
			"range": {
				"start": {
					"line": position.line,
					"character": position.character - count(last_word),
				},
				"end": position,
			},
			"newText": ref,
		},
		"_regal": {"provider": "rulerefs"},
	}
]
