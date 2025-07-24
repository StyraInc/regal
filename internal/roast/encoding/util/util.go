package util

import (
	jsoniter "github.com/json-iterator/go"
)

func WriteValsArray[T any](stream *jsoniter.Stream, vals []T) {
	stream.WriteArrayStart()

	for i, val := range vals {
		if i > 0 {
			stream.WriteMore()
		}

		stream.WriteVal(val)
	}

	stream.WriteArrayEnd()
}

func WriteStringsArray(stream *jsoniter.Stream, vals []string) {
	stream.WriteArrayStart()

	for i, val := range vals {
		if i > 0 {
			stream.WriteMore()
		}

		stream.WriteString(val)
	}

	stream.WriteArrayEnd()
}

func WriteObject[V any](stream *jsoniter.Stream, obj map[string]V) {
	stream.WriteObjectStart()

	i := 0

	for key, value := range obj {
		if i > 0 {
			stream.WriteMore()
		}

		stream.WriteObjectField(key)
		stream.WriteVal(value)

		i++
	}

	stream.WriteObjectEnd()
}
