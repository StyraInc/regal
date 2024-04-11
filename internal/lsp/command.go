package lsp

func toAnySlice(a []string) []any {
	b := make([]any, len(a))
	for i := range a {
		b[i] = a[i]
	}

	return b
}

func FmtCommand(args []string) Command {
	return Command{
		Title:     "Format using opa-fmt",
		Command:   "regal.fmt",
		Tooltip:   "Format using opa-fmt",
		Arguments: toAnySlice(args),
	}
}

func FmtV1Command(args []string) Command {
	return Command{
		Title:     "Format for Rego v1 using opa-fmt",
		Command:   "regal.fmt.v1",
		Tooltip:   "Format for Rego v1 using opa-fmt",
		Arguments: toAnySlice(args),
	}
}
