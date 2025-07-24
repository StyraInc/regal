package encoding

import (
	"sync"
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

type setCodec struct{}

func (*setCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

type set struct {
	elems     map[int]*ast.Term
	keys      []*ast.Term
	hash      int
	ground    bool
	sortGuard *sync.Once
}

func (*setCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	s := *((*set)(ptr))

	writeTermsArray(stream, s.keys)
}
