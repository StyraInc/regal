//nolint:gochecknoglobals,wrapcheck
package builtins

import (
	"errors"

	"github.com/anderseknert/roast/pkg/encoding"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/tester"
	"github.com/open-policy-agent/opa/topdown/builtins"
	"github.com/open-policy-agent/opa/types"

	"github.com/styrainc/regal/internal/parse"
)

// RegalParseModuleMeta metadata for regal.parse_module.
var RegalParseModuleMeta = &rego.Function{
	Name: "regal.parse_module",
	Decl: types.NewFunction(
		types.Args(
			types.Named("filename", types.S).Description("file name to attach to AST nodes' locations"),
			types.Named("rego", types.S).Description("Rego module"),
		),
		types.Named("output", types.NewObject(nil, types.NewDynamicProperty(types.S, types.A))),
	),
}

// RegalLastMeta metadata for regal.last.
var RegalLastMeta = &rego.Function{
	Name: "regal.last",
	Decl: types.NewFunction(
		types.Args(
			types.Named("array", types.NewArray(nil, types.A)).
				Description("performance optimized last index retrieval"),
		),
		types.Named("element", types.A),
	),
}

// RegalParseModule regal.parse_module, like rego.parse_module but with location data included in AST.
func RegalParseModule(_ rego.BuiltinContext, filename *ast.Term, policy *ast.Term) (*ast.Term, error) {
	policyStr, err := builtins.StringOperand(policy.Value, 1)
	if err != nil {
		return nil, err
	}

	filenameStr, err := builtins.StringOperand(filename.Value, 2)
	if err != nil {
		return nil, err
	}

	module, err := ast.ParseModuleWithOpts(string(filenameStr), string(policyStr), parse.ParserOptions())
	if err != nil {
		return nil, err
	}

	enhancedAST, err := parse.PrepareAST(string(filenameStr), string(policyStr), module)
	if err != nil {
		return nil, err
	}

	roast, err := encoding.JSON().MarshalToString(enhancedAST)
	if err != nil {
		return nil, err
	}

	term, err := ast.ParseTerm(roast)
	if err != nil {
		return nil, err
	}

	return term, nil
}

var errAiob = errors.New("array index out of bounds")

// RegalLast regal.last returns the last element of an array.
func RegalLast(_ rego.BuiltinContext, arr *ast.Term) (*ast.Term, error) {
	arrOp, err := builtins.ArrayOperand(arr.Value, 1)
	if err != nil {
		return nil, err
	}

	if arrOp.Len() == 0 {
		return nil, errAiob
	}

	return arrOp.Elem(arrOp.Len() - 1), nil
}

// TestContextBuiltins returns the list of builtins as expected by the test runner.
func TestContextBuiltins() []*tester.Builtin {
	return []*tester.Builtin{
		{
			Decl: &ast.Builtin{
				Name: RegalParseModuleMeta.Name,
				Decl: RegalParseModuleMeta.Decl,
			},
			Func: rego.Function2(RegalParseModuleMeta, RegalParseModule),
		},
		{
			Decl: &ast.Builtin{
				Name: RegalLastMeta.Name,
				Decl: RegalLastMeta.Decl,
			},
			Func: rego.Function1(RegalLastMeta, RegalLast),
		},
	}
}
