package compile

import (
	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/pkg/builtins"
)

func NewCompilerWithRegalBuiltins() *ast.Compiler {
	caps := ast.CapabilitiesForThisVersion()
	caps.Builtins = append(caps.Builtins, &ast.Builtin{
		Name: builtins.RegalParseModuleMeta.Name,
		Decl: builtins.RegalParseModuleMeta.Decl,
	})
	caps.Builtins = append(caps.Builtins, &ast.Builtin{
		Name: builtins.RegalJSONPrettyMeta.Name,
		Decl: builtins.RegalJSONPrettyMeta.Decl,
	})
	caps.Builtins = append(caps.Builtins, &ast.Builtin{
		Name: builtins.RegalLastMeta.Name,
		Decl: builtins.RegalLastMeta.Decl,
	})

	return ast.NewCompiler().WithCapabilities(caps)
}
