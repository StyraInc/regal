package lsp

import (
	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/lsp/types"
)

type BuiltInCall struct {
	Builtin  *ast.Builtin
	Location *ast.Location
	Args     []*ast.Term
}

func positionFromLocation(loc *ast.Location) types.Position {
	return types.Position{
		Line:      uint(loc.Row - 1),
		Character: uint(loc.Col - 1),
	}
}

// AllBuiltinCalls returns all built-in calls in the module, excluding operators
// and any other function identified by an infix.
func AllBuiltinCalls(module *ast.Module) []BuiltInCall {
	builtinCalls := make([]BuiltInCall, 0)

	callVisitor := ast.NewGenericVisitor(func(x interface{}) bool {
		var terms []*ast.Term

		switch node := x.(type) {
		case ast.Call:
			terms = node
		case *ast.Expr:
			if call, ok := node.Terms.([]*ast.Term); ok {
				terms = call
			}
		default:
			return false
		}

		if len(terms) == 0 {
			return false
		}

		if b, ok := builtins[terms[0].Value.String()]; ok {
			// Exclude operators and similar builtins
			if b.Infix != "" {
				return false
			}

			builtinCalls = append(builtinCalls, BuiltInCall{
				Builtin:  b,
				Location: terms[0].Location,
				Args:     terms[1:],
			})
		}

		return false
	})

	callVisitor.Walk(module)

	return builtinCalls
}
