package lsp

import (
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/types"
)

func createInlayTooltip(named *types.NamedType) string {
	if named.Descr == "" {
		return fmt.Sprintf("Type: `%s`", named.Type.String())
	}

	return fmt.Sprintf("%s\n\nType: `%s`", named.Descr, named.Type.String())
}

func getInlayHints(module *ast.Module) []InlayHint {
	inlayHints := make([]InlayHint, 0)

	for _, call := range AllBuiltinCalls(module) {
		for i, arg := range call.Builtin.Decl.NamedFuncArgs().Args {
			if len(call.Args) <= i {
				// avoid panic if provided a builtin function where the args
				// have yet to be provided, like if the user types `split()`
				continue
			}

			if named, ok := arg.(*types.NamedType); ok {
				inlayHints = append(inlayHints, InlayHint{
					Position:     positionFromLocation(call.Args[i].Location),
					Label:        named.Name + ":",
					Kind:         2,
					PaddingLeft:  false,
					PaddingRight: true,
					Tooltip: MarkupContent{
						Kind:  "markdown",
						Value: createInlayTooltip(named),
					},
				})
			}
		}
	}

	return inlayHints
}
