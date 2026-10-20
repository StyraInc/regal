# METADATA
# description: Common functions for finding AST elements at a given position
# schemas:
#   - input: schema.regal.lsp.common
package regal.lsp.util.find

import data.regal.ast
import data.regal.lsp.location
import data.regal.lsp.util.range

# METADATA
# description: find the function argument at the given position, if any
# schemas:
#   - input.params: schema.regal.lsp.textdocumentposition
arg_at_position := [arg, rule_index] if {
	text := input.regal.file.lines[input.params.position.line]

	# `+ 1` converts LSP 0-based char position to word_at's 1-based column
	word := location.word_at(text, input.params.position.character + 1)

	some rule_index
	arg := data.workspace.parsed[input.params.textDocument.uri].rules[rule_index].head.args[_]

	arg.type == "var"
	arg.value == word.text

	arg_pos := range.parse(arg.location)

	input.params.position.line == arg_pos.start.line
	input.params.position.character >= arg_pos.start.character
	input.params.position.character <= arg_pos.end.character
}

# METADATA
# description: find the rule at input.params.position if any
# schemas:
#   - input.params: schema.regal.lsp.textdocumentposition
rule_at_position := [rule, rule_index] if {
	some rule_index, rule in data.workspace.parsed[input.params.textDocument.uri].rules

	range.contains_position(range.parse(rule.location), input.params.position)
}

# METADATA
# description: find the import statement at input.params.position, if any
# schemas:
#   - input.params: schema.regal.lsp.textdocumentposition
import_at_position := imp if {
	some imp in data.workspace.parsed[input.params.textDocument.uri].imports

	range.contains_position(range.parse(imp.path.location), input.params.position)
}

# METADATA
# description: |
#   find the import statement at input.params.position, if the position is on its alias.
#   The alias has no location in the AST, so it is found in the text following the import path.
# schemas:
#   - input.params: schema.regal.lsp.textdocumentposition
import_alias_at_position := imp if {
	some imp in data.workspace.parsed[input.params.textDocument.uri].imports

	imp.alias

	path_end := range.parse(imp.path.location).end

	input.params.position.line == path_end.line

	rest := substring(input.regal.file.lines[path_end.line], path_end.character, -1)
	start := path_end.character + count(regex.find_n(`^\s+as\s+`, rest, 1)[0])

	input.params.position.character >= start
	input.params.position.character <= start + count(imp.alias)
}

# METADATA
# description: |
#   find the `some`-declared variable at the given position, if any.
# schemas:
#   - input.params: schema.regal.lsp.textdocumentposition
some_var_at_position := [var, rule_index] if {
	text := input.regal.file.lines[input.params.position.line]

	# `+ 1` converts LSP 0-based char position to word_at's 1-based column
	word := location.word_at(text, input.params.position.character + 1)

	some kind in ["some", "somein"]
	some rule_index, vars in ast.found.vars
	some var in vars[kind]

	var.value == word.text

	var_pos := range.parse(var.location)
	input.params.position.line == var_pos.start.line
	input.params.position.character >= var_pos.start.character
	input.params.position.character <= var_pos.end.character
}

# METADATA
# description: |
#   find the `every`-declared variable at the given position, if any.
# schemas:
#   - input.params: schema.regal.lsp.textdocumentposition
every_var_at_position := [var, every_terms] if {
	text := input.regal.file.lines[input.params.position.line]

	# `+ 1` converts LSP 0-based char position to word_at's 1-based column
	word := location.word_at(text, input.params.position.character + 1)

	some every_blocks in ast.found.every
	some every_terms in every_blocks
	some kind in ["key", "value"]

	var := every_terms[kind]
	var.type == "var"
	var.value == word.text

	var_pos := range.parse(var.location)
	input.params.position.line == var_pos.start.line
	input.params.position.character >= var_pos.start.character
	input.params.position.character <= var_pos.end.character
}

# METADATA
# description: |
#   find the variable declared inside a comprehension at the
#   given position, if any.
# schemas:
#   - input.params: schema.regal.lsp.textdocumentposition
comprehension_var_at_position := [var, comp] if {
	text := input.regal.file.lines[input.params.position.line]

	# `+ 1` converts LSP 0-based char position to word_at's 1-based column
	word := location.word_at(text, input.params.position.character + 1)

	some comps in ast.found.comprehensions
	some comp in comps
	some var in _comp_declared_vars(comp.value.body)

	var.value == word.text

	var_pos := range.parse(var.location)
	input.params.position.line == var_pos.start.line
	input.params.position.character >= var_pos.start.character
	input.params.position.character <= var_pos.end.character
}

_comp_declared_vars(body) := [v |
	some expr in body
	some symbol in expr.terms.symbols
	some v in array.slice(symbol.value, 1, 100)
	v.type == "var"
	not startswith(v.value, "$")
]
