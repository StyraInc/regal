package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

type termCodec struct{}

func (*termCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*termCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	term := *((*ast.Term)(ptr))

	stream.WriteObjectStart()

	if term.Location != nil {
		stream.WriteObjectField(strLocation)
		stream.WriteVal(term.Location)
	}

	if term.Value != nil {
		if term.Location != nil {
			stream.WriteMore()
		}

		stream.WriteObjectField(strType)
		stream.WriteString(ast.ValueName(term.Value))
		stream.WriteMore()
		stream.WriteObjectField(strValue)
		stream.WriteVal(term.Value)
	}

	stream.WriteObjectEnd()
}
