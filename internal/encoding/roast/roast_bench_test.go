package roast

import (
	"encoding/json/v2"
	"io"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
)

// 262180 ns/op      404445 B/op      2648 allocs/op // jsoniter
// 177527 ns/op          50 B/op         1 allocs/op // json/v2
func BenchmarkSerializeModule(b *testing.B) {
	policy := mustReadTestFile(b, "testdata/policy.rego")
	module := ast.MustParseModuleWithOpts(string(policy), ast.ParserOptions{ProcessAnnotation: true})

	var jsonResult []byte

	b.Run("json/v2", func(b *testing.B) {
		opts := json.JoinOptions(Options)
		for b.Loop() {
			if err := json.MarshalWrite(io.Discard, module, opts); err != nil {
				b.Fatalf("failed to marshal module: %v", err)
			}
		}
	})

	_ = jsonResult
}

// 2658 ns/op     288 B/op      17 allocs/op
// 2508 ns/op     224 B/op      13 allocs/op
func BenchmarkAnnotationsEncoding(b *testing.B) {
	annotations := annotations()

	for b.Loop() {
		_ = json.MarshalWrite(io.Discard, annotations, Options)
	}
}
