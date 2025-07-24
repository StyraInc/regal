package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

type withCodec struct{}

func (*withCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*withCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	with := *((*ast.With)(ptr))

	stream.WriteObjectStart()

	if with.Location != nil {
		stream.WriteObjectField(strLocation)
		stream.WriteVal(with.Location)
		stream.WriteMore()
	}

	stream.WriteObjectField(strTarget)
	stream.WriteVal(with.Target)
	stream.WriteMore()
	stream.WriteObjectField(strValue)
	stream.WriteVal(with.Value)

	stream.WriteObjectEnd()
}
