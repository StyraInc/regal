package util

import (
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
)

// UnquotedPath returns a slice of strings from a path without quotes.
// e.g. data.foo["bar"] -> ["foo", "bar"], note that the data is not included.
func UnquotedPath(path ast.Ref) []string {
	ret := make([]string, 0, len(path)-1)
	for _, ref := range path[1:] {
		ret = append(ret, strings.Trim(ref.String(), `"`))
	}

	return ret
}
