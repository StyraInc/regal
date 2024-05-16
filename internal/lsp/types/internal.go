package types

import "github.com/open-policy-agent/opa/ast"

type BuiltinPosition struct {
	Builtin *ast.Builtin
	Line    uint
	Start   uint
	End     uint
}
