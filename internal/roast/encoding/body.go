package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/open-policy-agent/regal/internal/roast/encoding/util"
)

type bodyCodec struct{}

func (*bodyCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*bodyCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	body := *((*ast.Body)(ptr))

	util.WriteValsArray(stream, body)
}
