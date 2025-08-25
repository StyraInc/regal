package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/open-policy-agent/regal/internal/roast/encoding/util"
)

type exprCodec struct{}

func (*exprCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*exprCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	expr := *((*ast.Expr)(ptr))

	stream.WriteObjectStart()

	hasWritten := false

	if expr.Location != nil {
		stream.WriteObjectField(strLocation)
		stream.WriteVal(expr.Location)

		hasWritten = true
	}

	if expr.Negated {
		if hasWritten {
			stream.WriteMore()
		}

		stream.WriteObjectField(strNegated)
		stream.WriteBool(expr.Negated)

		hasWritten = true
	}

	if expr.Generated {
		if hasWritten {
			stream.WriteMore()
		}

		stream.WriteObjectField(strGenerated)
		stream.WriteBool(expr.Generated)

		hasWritten = true
	}

	if len(expr.With) > 0 {
		if hasWritten {
			stream.WriteMore()
		}

		stream.WriteObjectField(strWith)
		util.WriteValsArray(stream, expr.With)

		hasWritten = true
	}

	if expr.Terms != nil {
		if hasWritten {
			stream.WriteMore()
		}

		stream.WriteObjectField(strTerms)

		switch t := expr.Terms.(type) {
		case *ast.Term:
			stream.WriteVal(t)
		case []*ast.Term:
			writeTermsArray(stream, t)
		case *ast.SomeDecl:
			stream.WriteVal(t)
		case *ast.Every:
			stream.WriteVal(t)
		}
	}

	stream.WriteObjectEnd()
}
