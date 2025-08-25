package encoding

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/open-policy-agent/regal/internal/roast/encoding/util"
)

type annotationsCodec struct{}

func (*annotationsCodec) IsEmpty(_ unsafe.Pointer) bool {
	return false
}

func (*annotationsCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	a := *((*ast.Annotations)(ptr))

	stream.WriteObjectStart()

	if a.Location != nil {
		stream.WriteObjectField(strLocation)
		stream.WriteVal(a.Location)
		stream.WriteMore()
	}

	stream.WriteObjectField(strScope)
	stream.WriteString(a.Scope)

	if a.Title != "" {
		stream.WriteMore()
		stream.WriteObjectField(strTitle)
		stream.WriteString(a.Title)
	}

	if a.Description != "" {
		stream.WriteMore()
		stream.WriteObjectField(strDescription)
		stream.WriteString(a.Description)
	}

	if a.Entrypoint {
		stream.WriteMore()
		stream.WriteObjectField(strEntrypoint)
		stream.WriteBool(a.Entrypoint)
	}

	if len(a.Organizations) > 0 {
		stream.WriteMore()
		stream.WriteObjectField(strOrganizations)
		util.WriteStringsArray(stream, a.Organizations)
	}

	if len(a.RelatedResources) > 0 {
		stream.WriteMore()
		stream.WriteObjectField(strRelatedResources)
		util.WriteValsArray(stream, a.RelatedResources)
	}

	if len(a.Authors) > 0 {
		stream.WriteMore()
		stream.WriteObjectField(strAuthors)
		util.WriteValsArray(stream, a.Authors)
	}

	if len(a.Schemas) > 0 {
		stream.WriteMore()
		stream.WriteObjectField(strSchemas)
		util.WriteValsArray(stream, a.Schemas)
	}

	if len(a.Custom) > 0 {
		stream.WriteMore()
		stream.WriteObjectField(strCustom)
		util.WriteObject(stream, a.Custom)
	}

	stream.WriteObjectEnd()
}
