package compile

import (
	"io/fs"
	"os"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/embeds"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/builtins"
	"github.com/styrainc/regal/pkg/roast/encoding"
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

	_ = fs.WalkDir(embeds.SchemasFS, "schemas", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			return nil
		}

		var schemaAny any

		util.Must0(encoding.JSON().Unmarshal(util.Must(embeds.SchemasFS.ReadFile(path)), &schemaAny))

		spl := strings.Split(strings.TrimSuffix(path, ".json"), string(os.PathSeparator))
		ref := ast.Ref([]*ast.Term{ast.SchemaRootDocument}).Extend(ast.MustParseRef(strings.Join(spl[1:], ".")))

		schemaSet.Put(ref, schemaAny)

		return nil
	})

	return schemaSet
}

func NewCompilerWithRegalBuiltins() *ast.Compiler {
	return ast.NewCompiler().WithCapabilities(Capabilities())
}
