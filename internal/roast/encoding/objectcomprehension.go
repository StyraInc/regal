package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

type objectComprehensionCodec struct{}

func (*objectComprehensionCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*objectComprehensionCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	oc := *((*ast.ObjectComprehension)(ptr))

	stream.WriteObjectStart()

	stream.WriteObjectField(strKey)
	stream.WriteVal(oc.Key)
	stream.WriteMore()
	stream.WriteObjectField(strValue)
	stream.WriteVal(oc.Value)
	stream.WriteMore()
	stream.WriteObjectField(strBody)
	stream.WriteVal(oc.Body)

	stream.WriteObjectEnd()
}
