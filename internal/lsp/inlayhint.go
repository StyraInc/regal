package lsp

import (
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/types"

	"github.com/styrainc/regal/internal/lsp/rego"
	types2 "github.com/styrainc/regal/internal/lsp/types"
)

func createInlayTooltip(named *types.NamedType) string {
	if named.Descr == "" {
		return fmt.Sprintf("Type: `%s`", named.Type.String())
	}

	return fmt.Sprintf("%s\n\nType: `%s`", named.Descr, named.Type.String())
}

func getInlayHints(module *ast.Module) []types2.InlayHint {
	inlayHints := make([]types2.InlayHint, 0)

	for _, call := range rego.AllBuiltinCalls(module) {
		for i, arg := range call.Builtin.Decl.NamedFuncArgs().Args {
			if len(call.Args) <= i {
				// avoid panic if provided a builtin function where the args
				// have yet to be provided, like if the user types `split()`
				continue
			}

			if named, ok := arg.(*types.NamedType); ok {
				inlayHints = append(inlayHints, types2.InlayHint{
					Position:     rego.PositionFromLocation(call.Args[i].Location),
					Label:        named.Name + ":",
					Kind:         2,
					PaddingLeft:  false,
					PaddingRight: true,
					Tooltip: types2.MarkupContent{
						Kind:  "markdown",
						Value: createInlayTooltip(named),
					},
				})
			}
		}
	}

	return inlayHints
}
