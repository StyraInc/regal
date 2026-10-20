package regal.lsp.references

import data.regal.ast
import data.regal.lsp.util.find
import data.regal.lsp.util.range

# METADATA
# description: |
#   References to the import at the cursor within the same file, i.e. every use of its identifier.
_import_refs contains ref if {
	_import.path.value[0].value in {"input", "data"}
	identifier := ast.imported_identifier(_import)

	some i, vars in ast.found.var_terms
	some var in vars

	var.value == identifier
	not _declares_variable(i, identifier)

	ref := {
		"uri": input.params.textDocument.uri,
		"range": range.parse(var.location),
	}
}

_import_refs contains ref if {
	_include_declaration

	ref := {
		"uri": input.params.textDocument.uri,
		"range": range.parse(_import.path.location),
	}
}

_declares_variable(rule_index, name) if {
	some kind in ["args", "assign", "some", "somein", "every"]
	ast.found.vars[rule_index][kind][_].value == name
}

_import := find.import_at_position

_import := find.import_alias_at_position
