package encoding

import (
	"bytes"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

type locationCodec struct{}

var newLine = []byte("\n")

var sbPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

func (*locationCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*locationCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	location := *((*ast.Location)(ptr))

	var endRow, endCol int
	if location.Text == nil {
		endRow = location.Row
		endCol = location.Col
	} else {
		lines := bytes.Split(location.Text, newLine)

		numLines := len(lines)

		endRow = location.Row + numLines - 1

		if numLines == 1 {
			endCol = location.Col + len(location.Text)
		} else {
			lastLine := lines[numLines-1]
			endCol = len(lastLine) + 1
		}
	}

	sb := sbPool.Get().(*strings.Builder) //nolint:forcetypeassert

	sb.WriteString(strconv.Itoa(location.Row))
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(location.Col))
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(endRow))
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(endCol))

	stream.WriteString(sb.String())

	sb.Reset()
	sbPool.Put(sb)
}
