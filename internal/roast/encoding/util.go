package encoding

import (
	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

func writeTermsArray(stream *jsoniter.Stream, items []*ast.Term) {
	stream.WriteArrayStart()

	for i, item := range items {
		if i > 0 {
			stream.WriteMore()
		}

		stream.WriteVal(item)
	}

	stream.WriteArrayEnd()
}
