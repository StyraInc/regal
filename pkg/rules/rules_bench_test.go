package rules_test

import (
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/open-policy-agent/regal/internal/test/must"
	"github.com/open-policy-agent/regal/pkg/config"
	"github.com/open-policy-agent/regal/pkg/rules"
)

// BenchmarkInputFromPaths/without_versions_map-16       57  20672270 ns/op    67962876 B/op    1072698 allocs/op
// BenchmarkInputFromPaths/with_versions_map-16          56  19535831 ns/op    67945825 B/op    1073711 allocs/op
func BenchmarkInputFromPaths(b *testing.B) {
	tests := []struct {
		name  string
		vsmap map[string]ast.RegoVersion
	}{
		{name: "without versions map", vsmap: nil},
		{name: "with versions map", vsmap: map[string]ast.RegoVersion{"bundle": ast.RegoV1}},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			paths := bundleDirPaths(b)
			for b.Loop() {
				if _, err := rules.InputFromPaths(paths, "", tc.vsmap); err != nil {
					b.Fatalf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func bundleDirPaths(b *testing.B) []string {
	b.Helper()

	bundle := []string{"../../bundle"}
	ignore := []string{}

	return must.Return(config.FilterIgnoredPaths(bundle, ignore, true, ""))(b)
}
