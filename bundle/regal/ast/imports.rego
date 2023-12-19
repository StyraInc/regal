package regal.ast

import rego.v1

default imports := []

# METADATA
# description: |
#   same as input.imports but with a default value (`[]`), making
#   it safe to refer to without risk of halting evaluation
imports := input.imports

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

_arr(xs) := [y.value | some y in xs.path.value]

_has_keyword(["future", "keywords"], _)

_has_keyword(["future", "keywords", keyword], keyword)

_has_keyword(["rego", "v1"], _)
