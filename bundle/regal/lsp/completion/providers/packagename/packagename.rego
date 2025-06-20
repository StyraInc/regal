# METADATA
# description: |
#   the `packagename` providers suggests completions for package
#   name based on the directory structure whre the file is located
package regal.lsp.completion.providers.packagename

import data.regal.lsp.completion.kind
import data.regal.lsp.completion.location

# METADATA
# description: set of suggested package names
items contains item if {
	position := location.to_position(input.regal.context.location)
	line := input.regal.file.lines[position.line]

	startswith(line, "package ")
	position.character > 7

	ps := input.regal.environment.path_separator

	abs_dir := _base(input.regal.file.name)
	rel_dir := trim_prefix(abs_dir, input.regal.context.workspace_root)
	fix_dir := replace(replace(trim_prefix(rel_dir, ps), ".", "_"), ps, ".")

	word := location.ref_at(line, input.regal.context.location.col)

	some suggestion in _suggestions(fix_dir, word.text)

	item := {
		"label": suggestion,
		"kind": kind.folder,
		"detail": "suggested package name based on directory structure",
		"textEdit": {
			"range": location.word_range(word, position),
			"newText": concat("", [suggestion, "\n\n"]),
		},
	}
}

_base(path) := substring(path, 0, regal.last(indexof_n(path, "/")))

_suggestions(dir, text) := [path |
	parts := split(dir, ".")
	len_p := count(parts)
	some n in numbers.range(0, len_p)

	path := concat(".", array.slice(parts, n, len_p))
	path != ""

	startswith(path, text)
]
