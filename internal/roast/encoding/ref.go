package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

type refCodec struct{}

func (*refCodec) IsEmpty(ptr unsafe.Pointer) bool {
	ref := *((*ast.Ref)(ptr))

	return len(ref) == 0
}

func (*refCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	ref := *((*ast.Ref)(ptr))

	writeTermsArray(stream, ref)
}
