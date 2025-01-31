//nolint:gochecknoglobals,wrapcheck
package builtins

import (
	"strings"

	"github.com/anderseknert/roast/pkg/encoding"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/tester"
	"github.com/open-policy-agent/opa/v1/topdown/builtins"
	"github.com/open-policy-agent/opa/v1/types"

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
	filenameValue, err := builtins.StringOperand(filename.Value, 1)
	if err != nil {
		return nil, err
	}

	policyValue, err := builtins.StringOperand(policy.Value, 2)
	if err != nil {
		return nil, err
	}

	filenameStr := string(filenameValue)
	policyStr := string(policyValue)

	opts := parse.ParserOptions()

	// Allow testing Rego v0 modules. We could provide a separate builtin for this,
	// but the need for this will likely diminish over time, so let's start simple.
	if strings.HasSuffix(filenameStr, "_v0.rego") {
		opts.RegoVersion = ast.RegoV0
	}

	module, err := ast.ParseModuleWithOpts(filenameStr, policyStr, opts)
	if err != nil {
		return nil, err
	}

	enhancedAST, err := parse.PrepareAST(filenameStr, policyStr, module)
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

// RegalLast regal.last returns the last element of an array.
func RegalLast(_ rego.BuiltinContext, arr *ast.Term) (*ast.Term, error) {
	arrOp, err := builtins.ArrayOperand(arr.Value, 1)
	if err != nil {
		return nil, err
	}

	if arrOp.Len() > 0 {
		return arrOp.Elem(arrOp.Len() - 1), nil
	}

	// index out of bounds, but returning an error allocates
	// and we have no use for this information anyway.
	return nil, nil //nolint:nilnil
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
