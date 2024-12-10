package compile

import (
	"github.com/anderseknert/roast/pkg/encoding"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/embeds"
	"github.com/styrainc/regal/internal/util"
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
			Name: builtins.RegalLastMeta.Name,
			Decl: builtins.RegalLastMeta.Decl,
		})

	return caps
}

func RegalSchemaSet() *ast.SchemaSet {
	schemaSet := ast.NewSchemaSet()
	astSchema := util.Must(embeds.SchemasFS.ReadFile("schemas/regal-ast.json"))
	aggSchema := util.Must(embeds.SchemasFS.ReadFile("schemas/aggregate.json"))

	var astSchemaAny, aggSchemaAny any

	util.Must0(encoding.JSON().Unmarshal(astSchema, &astSchemaAny))
	schemaSet.Put(ast.MustParseRef("schema.regal.ast"), astSchemaAny)

	util.Must0(encoding.JSON().Unmarshal(aggSchema, &aggSchemaAny))
	schemaSet.Put(ast.MustParseRef("schema.regal.aggregate"), aggSchemaAny)

	return schemaSet
}

func NewCompilerWithRegalBuiltins() *ast.Compiler {
	return ast.NewCompiler().WithCapabilities(Capabilities())
}
