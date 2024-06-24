package regal.lsp.completion.providers.packagename

import rego.v1

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

items contains item if {
	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	invoke_suggestion(line, position)

	ps := input.regal.context.path_separator

	abs_dir := base(input.regal.file.name)
	rel_dir := trim_prefix(abs_dir, input.regal.context.workspace_root)
	fix_dir := replace(replace(trim_prefix(rel_dir, ps), ".", "_"), ps, ".")

	word := location.ref_at(line, input.regal.context.location.col)

	some suggestion in suggestions(fix_dir, word)

	item := {
		"label": suggestion,
		"kind": kind.folder,
		"detail": "suggested package name based on directory structure",
		"textEdit": {
			"range": location.word_range(word, position),
			"newText": sprintf("%s\n\n", [suggestion]),
		},
	}
}

invoke_suggestion(line, position) if {
	startswith(line, "package ")
	position.character > 7
}

base(path) := substring(path, 0, regal.last(indexof_n(path, "/")))

suggestions(dir, word) := [path |
	parts := split(dir, ".")
	len_p := count(parts)
	some n in numbers.range(0, len_p)

	path := concat(".", array.slice(parts, n, len_p))
	path != ""

	startswith(path, word.text)
]
