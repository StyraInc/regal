package compile

import (
	"github.com/open-policy-agent/opa/ast"
	opaUtil "github.com/open-policy-agent/opa/util"

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

	schemaSet.Put(ast.MustParseRef("schema.regal.ast"), opaUtil.MustUnmarshalJSON(astSchema))

	aggregateSchema := util.Must(embeds.SchemasFS.ReadFile("schemas/aggregate.json"))

	schemaSet.Put(ast.MustParseRef("schema.regal.aggregate"), opaUtil.MustUnmarshalJSON(aggregateSchema))

	return schemaSet
}

func NewCompilerWithRegalBuiltins() *ast.Compiler {
	return ast.NewCompiler().WithCapabilities(Capabilities())
}
