package noast

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"io"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/open-policy-agent/regal/internal/test/must"
)

func BenchmarkEncodeValue(b *testing.B) {
	for _, test := range ValueTests {
		// 2 alloc regardless of type, incurred by json.MarshalWrite having to do
		// a reflective lookup to find the right marshaller. Hence why we call the
		// concrete MarshalJSONTo functions directly wherever we can.
		b.Run(test.name, func(b *testing.B) {
			for b.Loop() {
				must.Equal(b, nil, json.MarshalWrite(io.Discard, test.want, Options))
			}
		})

		b.Run(test.name+" ValueMarshalToFn", func(b *testing.B) {
			enc := jsontext.NewEncoder(io.Discard, Options)
			for b.Loop() {
				must.Equal(b, nil, ValueMarshalToFn(enc, test.want))
			}
		})
	}
}

func BenchmarkDecodeValue(b *testing.B) {
	for _, test := range ValueTests {
		b.Run(test.name, func(b *testing.B) {
			for b.Loop() {
				var v ast.Value

				_ = json.Unmarshal(test.json, &v, Options)
			}
		})

		b.Run(test.name+" OPA decode", func(b *testing.B) {
			for b.Loop() {
				_, _ = opaDecodeToValue(test.json)
			}
		})
	}
}
