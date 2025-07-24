package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

type packageCodec struct{}

func (*packageCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*packageCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	pkg := *((*ast.Package)(ptr))

	stream.WriteObjectStart()

	if pkg.Location != nil {
		stream.WriteObjectField(strLocation)
		stream.WriteVal(pkg.Location)
	}

	if pkg.Path != nil {
		if pkg.Location != nil {
			stream.WriteMore()
		}

		stream.WriteObjectField(strPath)

		// Make a copy to avoid data race
		// https://github.com/StyraInc/regal/issues/1167
		pathCopy := pkg.Path.Copy()

		// Omit location of "data" part of path, at it isn't present in code
		pathCopy[0].Location = nil

		stream.WriteVal(pathCopy)
	}

	if stream.Attachment != nil {
		stream.WriteMore()
		stream.WriteObjectField(strAnnotations)
		stream.WriteVal(stream.Attachment)
	}

	stream.WriteObjectEnd()
}
