package compile

import (
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/util"

	"github.com/styrainc/regal/pkg/builtins"
)

func Capabilities() *ast.Capabilities {
	caps := ast.CapabilitiesForThisVersion()
	caps.Builtins = append(caps.Builtins,
		&ast.Builtin{
			Name: builtins.RegalParseModuleMeta.Name,
			Decl: builtins.RegalParseModuleMeta.Decl,
		},
		&ast.Builtin{
			Name: builtins.RegalJSONPrettyMeta.Name,
			Decl: builtins.RegalJSONPrettyMeta.Decl,
		},
		&ast.Builtin{
			Name: builtins.RegalLastMeta.Name,
			Decl: builtins.RegalLastMeta.Decl,
		})

	return caps
}

func SchemaSet(s []byte) *ast.SchemaSet {
	schema := util.MustUnmarshalJSON(s)
	schemaSet := ast.NewSchemaSet()
	schemaSet.Put(ast.MustParseRef("schema.regal.ast"), schema)

	return schemaSet
}

func NewCompilerWithRegalBuiltins() *ast.Compiler {
	return ast.NewCompiler().WithCapabilities(Capabilities())
}
