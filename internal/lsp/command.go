package lsp

import (
	"github.com/styrainc/regal/internal/lsp/types"
)

type commandArgs struct {
	// Target is the URI of the document for which the command applies to
	Target string `json:"target"`

	// Optional arguments, command dependent
	// Diagnostic is the diagnostic that is to be fixed in the target
	Diagnostic *types.Diagnostic `json:"diagnostic,omitempty"`
	// QueryPath is the path of the query to evaluate
	QueryPath string `json:"path,omitempty"`
	// Row is the row within the file where the command was run from
	Row int `json:"row,omitempty"`
}
