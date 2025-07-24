package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

type importCodec struct{}

func (*importCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*importCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	imp := *((*ast.Import)(ptr))

	stream.WriteObjectStart()

	if imp.Location != nil {
		stream.WriteObjectField(strLocation)
		stream.WriteVal(imp.Location)
	}

	if imp.Path != nil {
		if imp.Location != nil {
			stream.WriteMore()
		}

		stream.WriteObjectField(strPath)
		stream.WriteVal(imp.Path)

		if imp.Alias != "" {
			stream.WriteMore()
			stream.WriteObjectField(strAlias)
			stream.WriteVal(imp.Alias)
		}
	}

	stream.WriteObjectEnd()
}
