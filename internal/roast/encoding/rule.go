package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/open-policy-agent/regal/internal/roast/encoding/util"
	"github.com/open-policy-agent/regal/pkg/roast/rast"
)

type ruleCodec struct{}

func (*ruleCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*ruleCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	rule := *((*ast.Rule)(ptr))

	stream.WriteObjectStart()

	hasWritten := false

	if rule.Location != nil {
		stream.WriteObjectField(strLocation)
		stream.WriteVal(rule.Location)

		hasWritten = true
	}

	if len(rule.Annotations) > 0 {
		if hasWritten {
			stream.WriteMore()
		}

		stream.WriteObjectField(strAnnotations)
		util.WriteValsArray(stream, rule.Annotations)

		hasWritten = true
	}

	if rule.Default {
		if hasWritten {
			stream.WriteMore()
		}

		stream.WriteObjectField(strDefault)
		stream.WriteBool(rule.Default)

		hasWritten = true
	}

	if rule.Head != nil {
		if hasWritten {
			stream.WriteMore()
		}

		stream.WriteObjectField(strHead)
		stream.WriteVal(rule.Head)
	}

	if !rast.IsBodyGenerated(&rule) {
		if hasWritten {
			stream.WriteMore()
		}

		stream.WriteObjectField(strBody)
		stream.WriteVal(rule.Body)
	}

	if rule.Else != nil {
		stream.WriteMore()
		stream.WriteObjectField(strElse)
		stream.WriteVal(rule.Else)
	}

	stream.WriteObjectEnd()
}
