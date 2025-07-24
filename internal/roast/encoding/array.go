package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

type arrayCodec struct{}

func (*arrayCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*arrayCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	arr := *((*ast.Array)(ptr))

	stream.WriteArrayStart()

	i := 0

	arr.Foreach(func(term *ast.Term) {
		if i > 0 {
			stream.WriteMore()
		}

		stream.WriteVal(term)

		i++
	})

	stream.WriteArrayEnd()
}
