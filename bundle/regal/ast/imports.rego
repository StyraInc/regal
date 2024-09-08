package regal.ast

import rego.v1

default imports := []

# METADATA
# description: |
#   same as input.imports but with a default value (`[]`), making
#   it safe to refer to without risk of halting evaluation
# scope: document
imports := input.imports

# METADATA
# description: |
#   set of all names imported in the input module, meaning commonly the last part of any
#   imported ref, like "bar" in "data.foo.bar", or an alias like "baz" in "data.foo.bar as baz".
imported_identifiers contains _imported_identifier(imp) if {
	some imp in imports

	imp.path.value[0].value in {"input", "data"}
	count(imp.path.value) > 1
}

# METADATA
# description: |
#   map of all imported paths in the input module, keyed by their identifier or "namespace"
resolved_imports[identifier] := path if {
	some identifier in imported_identifiers

	# this should really be just a 1:1 mapping, but until OPA 1.0 we cannot
	# trust that there are no duplicate imports, or imports shadowing other
	# imports, which may render a runtime error here if two paths are written
	# to the same identifier key ... simplify this post 1.0
	paths := [path |
		some imp in imports

		_imported_identifier(imp) == identifier

		path := [part.value | some part in imp.path.value]
	]

	path := paths[0]
}

# METADATA
# description: |
#   returns true if provided path (like ["data", "foo", "bar"]) is in the
#   list of imports (which is commonly ast.imports)
imports_has_path(imports, path) if {
	some imp in imports

	_arr(imp) == path
}

# METADATA
# description: |
#   returns whether a keyword is imported in the policy, either explicitly
#   like "future.keywords.if" or implicitly like "future.keywords" or "rego.v1"
imports_keyword(imports, keyword) if {
	some imp in imports

	_has_keyword(_arr(imp), keyword)
}

_imported_identifier(imp) := imp.alias

_imported_identifier(imp) := regal.last(imp.path.value).value if not imp.alias

_arr(xs) := [y.value | some y in xs.path.value]

_has_keyword(["future", "keywords"], _)

_has_keyword(["future", "keywords", "every"], "in")

_has_keyword(["future", "keywords", keyword], keyword)

_has_keyword(["rego", "v1"], _)
