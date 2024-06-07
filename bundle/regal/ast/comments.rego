package regal.ast

import rego.v1

# METADATA
# description: all comments in the input AST with their "Text" attribute base64 decoded
comments_decoded := [decoded |
	some comment in input.comments
	decoded := object.union(comment, {"Text": base64.decode(comment.Text)})
]

# METADATA
# description: |
#   an array of partitions, i.e. arrays containing all comments grouped by their "blocks"
comments["blocks"] := comment_blocks(comments_decoded)

# METADATA
# description: set of all the standard metadata attribute names, as provided by OPA
comments["metadata_attributes"] := {
	"scope",
	"title",
	"description",
	"related_resources",
	"authors",
	"organizations",
	"schemas",
	"entrypoint",
	"custom",
}

# METADATA
# description: |
#   map of all ignore directive comments, like ("# regal ignore:line-length")
#   found in input AST, indexed by the row they're at
ignore_directives[row] := rules if {
	some comment in comments_decoded
	text := trim_space(comment.Text)

	i := indexof(text, "regal ignore:")
	i != -1

	list := regex.replace(substring(text, i + 13, -1), `\s`, "")

	row := comment.Location.row + 1
	rules := split(list, ",")
}

# METADATA
# description: |
#   returns an array of partitions, i.e. arrays containing all comments
#   grouped by their "blocks".
comment_blocks(comments) := [partition |
	rows := [row |
		some comment in comments
		row := comment.Location.row
	]
	breaks := _splits(rows)

	some j, k in breaks
	partition := array.slice(
		comments,
		breaks[j - 1] + 1,
		k + 1,
	)
]

_splits(xs) := array.concat(
	array.concat(
		# [-1] ++ [ all indices where there's a step larger than one ] ++ [length of xs]
		# the -1 is because we're adding +1 in array.slice
		[-1],
		[i |
			some i in numbers.range(0, count(xs) - 1)
			xs[i + 1] != xs[i] + 1
		],
	),
	[count(xs)],
)
