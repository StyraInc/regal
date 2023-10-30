package compile

import (
	"github.com/open-policy-agent/opa/ast"
	opa_util "github.com/open-policy-agent/opa/util"

	"github.com/styrainc/regal/internal/embeds"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/builtins"
)

func DefaultCapabilities() *ast.Capabilities {
	return Capabilities(*ast.CapabilitiesForThisVersion())
}

func Capabilities(caps ast.Capabilities) *ast.Capabilities {
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

	return &caps
}

func RegalSchemaSet() *ast.SchemaSet {
	schemaSet := ast.NewSchemaSet()

	astSchema := util.Must(embeds.SchemasFS.ReadFile("schemas/regal-ast.json"))

	schemaSet.Put(ast.MustParseRef("schema.regal.ast"), opa_util.MustUnmarshalJSON(astSchema))

	aggregateSchema := util.Must(embeds.SchemasFS.ReadFile("schemas/aggregate.json"))

	schemaSet.Put(ast.MustParseRef("schema.regal.aggregate"), opa_util.MustUnmarshalJSON(aggregateSchema))

	return schemaSet
}

func NewCompilerWithRegalBuiltins() *ast.Compiler {
	return ast.NewCompiler().WithCapabilities(DefaultCapabilities())
}
