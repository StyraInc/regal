package util

import (
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
)

func BenchmarkStringRepeatMake(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_ = stringRepeatMake("test", 1000)
	}
}

func stringRepeatMake(s string, n int) []*ast.Term {
	sl := make([]*ast.Term, n)
	for i := range s {
		sl[i] = &ast.Term{Value: ast.String("test")}
	}

	return sl
}

// Without pre-allocating, this is more than twice as slow and results in 5 allocs/op.
// BenchmarkFilter/Filter-10    5919769    191.0 ns/op    224 B/op    1 allocs/op
// ...
func BenchmarkFilter(b *testing.B) {
	strings := []string{
		"foo", "bar", "baz", "qux", "quux", "corge", "grault", "garply", "waldo", "fred", "plugh", "xyzzy", "thud",
		"x", "y", "z", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "the", "lazy", "dog", "jumped", "over",
		"the", "quick", "brown", "fox",
	}

	pred := func(s string) bool {
		return len(s) > 3
	}

	b.Run("Filter", func(b *testing.B) {
		for range b.N {
			_ = Filter(strings, pred)
		}
	})
}
