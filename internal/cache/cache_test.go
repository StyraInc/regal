package cache

import (
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
)

// 2 allocs.
func BenchmarkPutNew(b *testing.B) {
	cache := NewBaseCache()
	ref := ast.MustParseRef("data.foo.bar.baz")
	value := ast.String("qux")

	for b.Loop() {
		cache.Put(ref, value)
	}
}

// 0 allocs.
func BenchmarkGet(b *testing.B) {
	cache := NewBaseCache()
	ref := ast.MustParseRef("data.foo.bar")
	value := ast.NewObject(ast.Item(ast.StringTerm("baz"), ast.StringTerm("qux")))
	cache.Put(ref, value)

	r := ast.MustParseRef("data.foo.bar.baz")

	for b.Loop() {
		_ = cache.Get(r)
	}
}
