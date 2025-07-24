package encoding

import (
	"sync"
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

type objectCodec struct{}

func (*objectCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

type object struct {
	elems     map[int]*objectElem
	keys      objectElemSlice
	ground    int
	hash      int
	sortGuard *sync.Once
}

type objectElem struct {
	key   *ast.Term
	value *ast.Term
	next  *objectElem //nolint:unused
}

type objectElemSlice []*objectElem

func (*objectCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	o := *((*object)(ptr))

	stream.WriteArrayStart()

	for i, node := range o.keys {
		if i > 0 {
			stream.WriteMore()
		}

		stream.WriteArrayStart()
		stream.WriteVal(node.key)
		stream.WriteMore()
		stream.WriteVal(node.value)
		stream.WriteArrayEnd()
	}

	stream.WriteArrayEnd()
}
