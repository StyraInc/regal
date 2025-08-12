package query

const (
	Keywords          = "data.regal.ast.keywords"
	RuleHeadLocations = "data.regal.ast.rule_head_locations"
	CodeLens          = "data.regal.lsp.codelens.lenses"
	CodeAction        = "data.regal.lsp.codeaction.actions"
	DocumentLink      = "data.regal.lsp.documentlink.items"
	DocumentHighlight = "data.regal.lsp.documenthighlight.items"
	Completion        = "data.regal.lsp.completion.items"
	SignatureHelp     = "data.regal.lsp.signaturehelp.result"
)

func AllQueries() []string {
	return []string{
		Keywords, RuleHeadLocations, CodeLens, CodeAction, DocumentLink, DocumentHighlight, Completion, SignatureHelp,
	}
}
