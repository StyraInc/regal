package regal.ast

import rego.v1

comments_decoded := [decoded |
	some comment in input.comments
	decoded := object.union(comment, {"Text": base64.decode(comment.Text)})
]

comments["blocks"] := comment_blocks(comments_decoded)

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
