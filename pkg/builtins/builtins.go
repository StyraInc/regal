//nolint:gochecknoglobals,wrapcheck
package builtins

import (
	"bytes"
	"encoding/json"
	"errors"

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

// RegalJSONPrettyMeta metadata for regal.json_pretty.
var RegalJSONPrettyMeta = &rego.Function{
	Name: "regal.json_pretty",
	Decl: types.NewFunction(
		types.Args(
			types.Named("data", types.A).Description("data to marshal to JSON in a pretty format"),
		),
		types.Named("output", types.S),
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

	enhancedAST, err := parse.EnhanceAST(string(filenameStr), string(policyStr), module)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&enhancedAST); err != nil {
		return nil, err
	}

	term, err := ast.ParseTerm(buf.String())
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

// RegalJSONPretty regal.json_pretty, like json.marshal but with pretty formatting.
func RegalJSONPretty(_ rego.BuiltinContext, data *ast.Term) (*ast.Term, error) {
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	return ast.StringTerm(string(encoded)), nil
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
				Name: RegalJSONPrettyMeta.Name,
				Decl: RegalJSONPrettyMeta.Decl,
			},
			Func: rego.Function1(RegalJSONPrettyMeta, RegalJSONPretty),
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
