package lsp

import (
	"encoding/json"

	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/util"
)

func toAnySlice(a ...string) *[]any {
	b := make([]any, len(a))
	for i := range a {
		b[i] = a[i]
	}

	return &b
}

type commandArgs struct {
	Target     string            `json:"target"`
	Diagnostic *types.Diagnostic `json:"diagnostic,omitempty"`
}

func FmtCommand(target string) types.Command {
	bs := util.Must(json.Marshal(commandArgs{
		Target: target,
	}))

	return types.Command{
		Title:     "Format using opa-fmt",
		Command:   "regal.fix.opa-fmt",
		Tooltip:   "Format using opa-fmt",
		Arguments: toAnySlice(string(bs)),
	}
}

func FmtV1Command(target string) types.Command {
	bs := util.Must(json.Marshal(commandArgs{
		Target: target,
	}))

	return types.Command{
		Title:     "Format for Rego v1 using opa-fmt",
		Command:   "regal.fix.use-rego-v1",
		Tooltip:   "Format for Rego v1 using opa-fmt",
		Arguments: toAnySlice(string(bs)),
	}
}

func UseAssignmentOperatorCommand(target string, diag types.Diagnostic) types.Command {
	bs := util.Must(json.Marshal(commandArgs{
		Target:     target,
		Diagnostic: &diag,
	}))

	return types.Command{
		Title:     "Replace = with := in assignment",
		Command:   "regal.fix.use-assignment-operator",
		Tooltip:   "Replace = with := in assignment",
		Arguments: toAnySlice(string(bs)),
	}
}

func NoWhiteSpaceCommentCommand(target string, diag types.Diagnostic) types.Command {
	bs := util.Must(json.Marshal(commandArgs{
		Target:     target,
		Diagnostic: &diag,
	}))

	return types.Command{
		Title:     "Format comment to have leading whitespace",
		Command:   "regal.fix.no-whitespace-comment",
		Tooltip:   "Format comment to have leading whitespace",
		Arguments: toAnySlice(string(bs)),
	}
}

func DirectoryStructureMismatchCommand(target string, diag types.Diagnostic) types.Command {
	bs := util.Must(json.Marshal(commandArgs{
		Target:     target,
		Diagnostic: &diag,
	}))

	return types.Command{
		Title:     "Fix directory structure / package path mismatch",
		Command:   "regal.fix.directory-package-mismatch",
		Tooltip:   "Fix directory structure / package path mismatch",
		Arguments: toAnySlice(string(bs)),
	}
}

func NonRawRegexPatternCommand(target string, diag types.Diagnostic) types.Command {
	bs := util.Must(json.Marshal(commandArgs{
		Target:     target,
		Diagnostic: &diag,
	}))

	return types.Command{
		Title:     "Replace \" with ` in regex pattern",
		Command:   "regal.fix.non-raw-regex-pattern",
		Tooltip:   "Replace \" with ` in regex pattern",
		Arguments: toAnySlice(string(bs)),
	}
}
