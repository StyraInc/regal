package process

import rego.v1

symbols := {"keywords": _keywords, "builtins": _builtins}

_keywords[name] := path if {
	some p in _pages
	p[0] == "keywords"

	name := p[1]

	path := concat("/", p)
}

_builtins[name] := path if {
	some p in _pages
	p[0] == "builtins"

	l := count(p)

	l == 2

	name := p[1]

	count({p |
		some p in _pages
		p[0] == "builtins"
		p[1] == name
	}) < 2

	path := concat("/", p)
}

_builtins[name] := path if {
	some p in _pages
	p[0] == "builtins"

	l := count(p)

	l > 2

	name := concat(
		".",
		[
			replace(p[1], "_", "."),
			concat(".", array.slice(p, 2, l)),
		],
	)

	path := concat("/", p)
}

_prefix := "https://docs.styra.com/opa/rego-by-example/"

_pages contains page if {
	some url in input.urlset.url

	startswith(url.loc, _prefix)

	page := split(trim_prefix(url.loc, _prefix), "/")
}
