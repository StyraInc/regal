package lsp

import "github.com/styrainc/regal/internal/lsp/types"

func toAnySlice(a []string) []any {
	b := make([]any, len(a))
	for i := range a {
		b[i] = a[i]
	}

	return b
}

func FmtCommand(args []string) types.Command {
	return types.Command{
		Title:     "Format using opa-fmt",
		Command:   "regal.fix.opa-fmt",
		Tooltip:   "Format using opa-fmt",
		Arguments: toAnySlice(args),
	}
}

func FmtV1Command(args []string) types.Command {
	return types.Command{
		Title:     "Format for Rego v1 using opa-fmt",
		Command:   "regal.fix.use-rego-v1",
		Tooltip:   "Format for Rego v1 using opa-fmt",
		Arguments: toAnySlice(args),
	}
}

func UseAssignmentOperatorCommand(args []string) types.Command {
	return types.Command{
		Title:     "Replace = with := in assignment",
		Command:   "regal.fix.use-assignment-operator",
		Tooltip:   "Replace = with := in assignment",
		Arguments: toAnySlice(args),
	}
}

func NoWhiteSpaceCommentCommand(args []string) types.Command {
	return types.Command{
		Title:     "Format comment to have leading whitespace",
		Command:   "regal.fix.no-whitespace-comment",
		Tooltip:   "Format comment to have leading whitespace",
		Arguments: toAnySlice(args),
	}
}
