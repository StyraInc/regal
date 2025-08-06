package io

import (
	"slices"
	"testing"

	"github.com/open-policy-agent/opa/v1/util/test"
)

func TestFindManifestLocations(t *testing.T) {
	t.Parallel()

	fs := map[string]string{
		"/.git":                          "",
		"/foo/bar/baz/.manifest":         "",
		"/foo/bar/qux/.manifest":         "",
		"/foo/bar/.regal/.manifest.yaml": "",
		"/node_modules/.manifest":        "",
	}

	test.WithTempFS(fs, func(root string) {
		locations, err := FindManifestLocations(root)
		if err != nil {
			t.Error(err)
		}

		if len(locations) != 2 {
			t.Errorf("expected 2 locations, got %d", len(locations))
		}

		expected := []string{"foo/bar/baz", "foo/bar/qux"}

		if !slices.Equal(locations, expected) {
			t.Errorf("expected %v, got %v", expected, locations)
		}
	})
}

func BenchmarkLoadRegalBundlePath(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := LoadRegalBundlePath("../../bundle")
		if err != nil {
			b.Fatal(err)
		}
	}
}
