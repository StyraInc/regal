package compile

import (
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/embeds"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/builtins"

	"github.com/styrainc/roast/pkg/encoding"
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
		},
		&ast.Builtin{
			Name: builtins.RegalIsFormattedMeta.Name,
			Decl: builtins.RegalIsFormattedMeta.Decl,
		},
	)

	return caps
}

// RegalSchemaSet returns a SchemaSet containing the Regal schemas embedded in the binary.
// Currently only used by the test command. Should we want to expand the use of this later,
// we'll probably want to only read the schemas relevant to the context.
func RegalSchemaSet() *ast.SchemaSet {
	schemaSet := ast.NewSchemaSet()

	for _, entry := range util.Must(embeds.SchemasFS.ReadDir("schemas")) {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		schemaFile := util.Must(embeds.SchemasFS.ReadFile("schemas/" + entry.Name()))

		var schemaAny any

		util.Must0(encoding.JSON().Unmarshal(schemaFile, &schemaAny))

		ref := ast.MustParseRef("schema.regal." + strings.TrimSuffix(entry.Name(), ".json"))

		schemaSet.Put(ref, schemaAny)
	}

	return schemaSet
}

func NewCompilerWithRegalBuiltins() *ast.Compiler {
	return ast.NewCompiler().WithCapabilities(Capabilities())
}
