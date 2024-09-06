package regal.lsp.completion.providers.rulerefs

import rego.v1

import data.regal.ast

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

ref_is_internal(ref) if contains(ref, "._")

position := location.to_position(input.regal.context.location)

line := input.regal.file.lines[position.line]

word := location.ref_at(line, input.regal.context.location.col)

workspace_rule_refs contains ref if {
	some refs in data.workspace.defined_refs
	some ref in refs
}

parsed_current_file := data.workspace.parsed[input.regal.file.uri]

current_file_package := ast.ref_to_string(parsed_current_file["package"].path)

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

	not ref_is_internal(ref)

	strings.any_prefix_match(ref, current_file_imports)
}

other_package_refs contains ref if {
	some ref in workspace_rule_refs

	not ref in imported_package_refs
	not ref in current_package_refs

	not ref_is_internal(ref)
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

	prefix := regex.replace(imported_package, `\.[^\.]+$`, "")
	pkg_ref := trim_prefix(ref, sprintf("%s.", [prefix]))
}

# from any other package
rule_ref_suggestions contains ref if some ref in other_package_refs

# also suggest the unimported packages themselves
# e.g. data.foo.rule will also generate data.foo as a suggestion
rule_ref_suggestions contains pkg if {
	some ref in other_package_refs

	pkg := regex.replace(ref, `\.[^\.]+$`, "")
}

matching_rule_ref_suggestions contains ref if {
	line != ""
	location.in_rule_body(line)

	# \W is used here to match ( in the case of func() := ..., as well as the space in the case of rule := ...
	first_word := regex.split(`\W+`, trim_space(line))[0]

	some ref in rule_ref_suggestions

	startswith(ref, word.text)

	# this is to avoid suggesting a recursive rule, e.g. rule := rule, or func() := func()
	ref != first_word
}

items contains item if {
	some ref in matching_rule_ref_suggestions

	item := {
		"label": ref,
		"kind": kind.variable,
		"detail": "reference",
		"textEdit": {
			"range": location.word_range(word, position),
			"newText": ref,
		},
	}
}
