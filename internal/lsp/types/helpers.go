package types

import (
	"strconv"
	"strings"
)

func NewCompletionParams(uri string, line, char uint, context *CompletionContext) CompletionParams {
	return CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: line, Character: char},
		Context:      context,
	}
}

func Markdown(value string) *MarkupContent {
	return &MarkupContent{Kind: "markdown", Value: value}
}

func RangeBetween[T1, T2, T3, T4 iuint](startLine T1, startCharacter T2, endLine T3, endCharacter T4) Range {
	return Range{
		Start: Position{Line: uint(startLine), Character: uint(startCharacter)},
		End:   Position{Line: uint(endLine), Character: uint(endCharacter)},
	}
}

func (r Range) String() string {
	var sb strings.Builder
	sb.WriteString(strconv.FormatUint(uint64(r.Start.Line), 10))
	sb.WriteString(":")
	sb.WriteString(strconv.FormatUint(uint64(r.Start.Character), 10))
	sb.WriteString(":")
	sb.WriteString(strconv.FormatUint(uint64(r.End.Line), 10))
	sb.WriteString(":")
	sb.WriteString(strconv.FormatUint(uint64(r.End.Character), 10))

	return sb.String()
}
