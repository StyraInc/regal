package types

import "github.com/open-policy-agent/opa/ast"

// Ref is a generic construct for an object found in a Rego module.
// Ref is designed to be used in completions and provides information
// relevant to the object with that operation in mind.
type Ref struct {
	Kind RefKind
	// Label is a identifier for the object. e.g. data.package.rule.
	Label string
	// Detail is a small amount of additional information about the object.
	Detail string
	// Description is a longer description of the object and uses Markdown formatting.
	Description string
}

// RefKind represents the kind of object that a Ref represents.
// This is intended to toggle functionality and which UI symbols to use.
type RefKind int

const (
	Package RefKind = iota + 1
	Rule
	ConstantRule
	Function
)

type BuiltinPosition struct {
	Builtin *ast.Builtin
	Line    uint
	Start   uint
	End     uint
}
