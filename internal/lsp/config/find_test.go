package config

import (
	"cmp"
	"slices"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/v1/util/test"
)

func TestFindConfigRoots(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		FS       map[string]string
		Expected []string
	}{
		"no config roots": {
			FS:       map[string]string{},
			Expected: []string{},
		},
		"single config root at root": {
			FS: map[string]string{
				".regal/config.yaml": "",
			},
			Expected: []string{"/"},
		},
		"single config root at root with .regal.yaml": {
			FS: map[string]string{
				".regal.yaml": "",
			},
			Expected: []string{"/"},
		},
		"two config roots, one higher": {
			FS: map[string]string{
				".regal/config.yaml": "",
				"foo/.regal.yaml":    "",
			},
			Expected: []string{
				"/",
				"/foo",
			},
		},
		"two config roots, one higher, not in root dir": {
			FS: map[string]string{
				"foo/.regal.yaml":            "",
				"bar/baz/.regal/config.yaml": "",
			},
			Expected: []string{
				"/bar/baz",
				"/foo",
			},
		},
		"two config roots, equal depth": {
			FS: map[string]string{
				"bar/.regal/config.yaml": "",
				"foo/.regal.yaml":        "",
			},
			Expected: []string{
				"/bar",
				"/foo",
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			test.WithTempFS(testData.FS, func(root string) {
				got, err := FindConfigRoots(root)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				gotTrimmed := make([]string, len(got))

				for i, path := range got {
					trimmed := cmp.Or(strings.TrimPrefix(path, root), "/")
					gotTrimmed[i] = trimmed
				}

				if !slices.Equal(gotTrimmed, testData.Expected) {
					t.Fatalf("Expected %v, got %v", testData.Expected, gotTrimmed)
				}
			})
		})
	}
}
