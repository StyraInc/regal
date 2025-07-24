package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

type everyCodec struct{}

func (*everyCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*everyCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	every := *((*ast.Every)(ptr))

	stream.WriteObjectStart()

	if every.Location != nil {
		stream.WriteObjectField(strLocation)
		stream.WriteVal(every.Location)
		stream.WriteMore()
	}

	stream.WriteObjectField(strKey)
	stream.WriteVal(every.Key)
	stream.WriteMore()

	stream.WriteObjectField(strValue)
	stream.WriteVal(every.Value)
	stream.WriteMore()

	stream.WriteObjectField(strDomain)
	stream.WriteVal(every.Domain)
	stream.WriteMore()

	stream.WriteObjectField(strBody)
	stream.WriteVal(every.Body)

	stream.WriteObjectEnd()
}
