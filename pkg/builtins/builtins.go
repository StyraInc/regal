//nolint:gochecknoglobals,wrapcheck
package builtins

import (
	"bytes"
	"encoding/json"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
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

// RegalJSONPretty regal.json_pretty, like json.marshal but with pretty formatting.
func RegalJSONPretty(_ rego.BuiltinContext, data *ast.Term) (*ast.Term, error) {
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	return ast.StringTerm(string(encoded)), nil
}
