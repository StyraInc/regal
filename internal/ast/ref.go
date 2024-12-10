package ast

import (
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
)

// RefToString converts an ast.Ref to a readable string, e.g. data.foo[bar].
func RefToString(ref ast.Ref) string {
	sb := strings.Builder{}

	for i, part := range ref {
		if part.IsGround() {
			if i > 0 {
				sb.WriteString(".")
			}

			sb.WriteString(strings.Trim(part.Value.String(), `"`))
		} else {
			if i == 0 {
				sb.WriteString(strings.Trim(part.Value.String(), `"`))
			} else {
				sb.WriteString("[")
				sb.WriteString(strings.Trim(part.Value.String(), `"`))
				sb.WriteString("]")
			}
		}
	}

	return sb.String()
}
